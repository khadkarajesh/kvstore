package server

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/rajesh/kvstore/proto"
	raft "github.com/rajesh/kvstore/raft"
)

// RaftServer implements the gRPC RaftServiceServer interface.
type RaftServer struct {
	pb.UnimplementedRaftServiceServer
	mu            sync.Mutex
	state         raft.NodeState
	currentTerm   int64
	votedFor      uint32
	lastLogIndex  uint64
	lastLogTerm   uint64
	electionTimer     *time.Timer
	timerResetCount   int
}

func (s *RaftServer) RequestVote(ctx context.Context, req *pb.RequestVoteRequest) (*pb.RequestVoteResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if req.Term < uint64(s.currentTerm) {
		return &pb.RequestVoteResponse{
			Term:        uint64(s.currentTerm),
			VoteGranted: false,
		}, nil
	}

	if req.Term > uint64(s.currentTerm) {
		s.currentTerm = int64(req.Term)
		s.votedFor = 0
		s.state = raft.Follower
	}

	// Check if we can vote: haven't voted OR already voted for this candidate
	alreadyVoted := s.votedFor != 0 && s.votedFor != req.CandidateId

	logUpToDate := req.LastLogTerm > s.lastLogTerm ||
		(req.LastLogTerm == s.lastLogTerm && req.LastLogIndex >= s.lastLogIndex)

	if !alreadyVoted && logUpToDate {
		s.votedFor = req.CandidateId
		return &pb.RequestVoteResponse{
			Term:        uint64(s.currentTerm),
			VoteGranted: true,
		}, nil
	}

	return &pb.RequestVoteResponse{
		Term:        uint64(s.currentTerm),
		VoteGranted: false,
	}, nil
}

func randomElectionTimeout() time.Duration {
	return time.Duration(150+rand.Intn(150)) * time.Millisecond
}

func (s *RaftServer) runElectionTimer() {
	for {
		<-s.electionTimer.C // blocks until timer fires

		s.mu.Lock()
		// timer fired → no heartbeat received in time
		// → transition to Candidate, start election
		s.state = raft.Candidate
		s.currentTerm++
		s.votedFor = 0
		s.mu.Unlock()
	}
}

func resetElectionTimer(s *RaftServer) {
	if !s.electionTimer.Stop() {
		select {
		case <-s.electionTimer.C:
		default:
		}
	}
	s.electionTimer.Reset(randomElectionTimeout())
	s.timerResetCount++
}

func (s *RaftServer) AppendEntries(ctx context.Context, req *pb.AppendEntriesRequest) (*pb.AppendEntriesResponse, error) {
	// Progressive Implementation Order

	// Step 1 — Heartbeat only
	//   Handle AppendEntries with empty entries[]
	//   Reset election timer on valid heartbeat
	//   Reject if req.Term < currentTerm
	//   Tests: heartbeat accepted, stale leader rejected, timer resets

	// Step 2 — Consistency check
	//   Implement prevLogIndex/prevLogTerm verification
	//   Return false if log doesn't match at prevLogIndex
	//   Tests: matching log accepts, mismatched log rejects

	// Step 3 — Append entries
	//   On successful consistency check, append new entries
	//   Truncate conflicting entries before appending
	//   Tests: entries appended, conflicts overwritten

	// Step 4 — Commit index
	//   Update commitIndex = min(leaderCommit, lastLogIndex)
	//   Apply committed entries to state machine
	//   Tests: commitIndex advances, state machine applies

	// Start with Step 1. Get heartbeat working with tests before touching
	// log replication. Each step has a clear set of test cases you can
	// derive from the rules above — same approach as RequestVote.

	s.mu.Lock()
	defer s.mu.Unlock()

	// This is a heartbeat. Reset election timer if valid.
	if req.Term < uint64(s.currentTerm) {
		return &pb.AppendEntriesResponse{
			Term:    uint64(s.currentTerm),
			Success: false,
		}, nil
	}

	if req.Term > uint64(s.currentTerm) {
		s.currentTerm = int64(req.Term)
		s.state = raft.Follower
	} else if req.Term == uint64(s.currentTerm) && s.state == raft.Candidate {
		s.state = raft.Follower
	}

	if len(req.Entries) == 0 {
		// Valid heartbeat, reset timer
		resetElectionTimer(s)
		return &pb.AppendEntriesResponse{
			Term:    uint64(s.currentTerm),
			Success: true,
		}, nil
	}

	return nil, status.Errorf(codes.Unimplemented, "AppendEntries not implemented")
}

// NewRaftServer creates and returns a new RaftServer instance.
func NewRaftServer() *RaftServer {
	s := &RaftServer{
		state:         raft.Follower,
		electionTimer: time.NewTimer(randomElectionTimeout()),
	}
	go s.runElectionTimer()
	return s
}

// Register wires the RaftServer into a gRPC server.
func (s *RaftServer) Register(grpcServer *grpc.Server) {
	pb.RegisterRaftServiceServer(grpcServer, s)
}
