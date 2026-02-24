package server

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"google.golang.org/grpc"

	pb "github.com/rajesh/kvstore/proto"
	raft "github.com/rajesh/kvstore/raft"
)

// RaftServer implements the gRPC RaftServiceServer interface.
type RaftServer struct {
	pb.UnimplementedRaftServiceServer
	mu              sync.Mutex
	state           raft.NodeState
	currentTerm     uint64
	votedFor        uint32
	lastHeartbeat   time.Time
	timerResetCount int
	log             []*pb.LogEntry
	commitIndex     uint64
}

func (s *RaftServer) RequestVote(ctx context.Context, req *pb.RequestVoteRequest) (*pb.RequestVoteResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if req.Term < s.currentTerm {
		return &pb.RequestVoteResponse{
			Term:        s.currentTerm,
			VoteGranted: false,
		}, nil
	}

	if req.Term > s.currentTerm {
		s.currentTerm = req.Term
		s.votedFor = 0
		s.state = raft.Follower
	}

	// Check if we can vote: haven't voted OR already voted for this candidate
	alreadyVoted := s.votedFor != 0 && s.votedFor != req.CandidateId

	logUpToDate := req.LastLogTerm > s.getLastLogTerm() ||
		(req.LastLogTerm == s.getLastLogTerm() && req.LastLogIndex >= s.getLastLogIndex())

	if !alreadyVoted && logUpToDate {
		s.votedFor = req.CandidateId
		return &pb.RequestVoteResponse{
			Term:        s.currentTerm,
			VoteGranted: true,
		}, nil
	}

	return &pb.RequestVoteResponse{
		Term:        s.currentTerm,
		VoteGranted: false,
	}, nil
}

func randomElectionTimeout() time.Duration {
	return time.Duration(150+rand.Intn(150)) * time.Millisecond
}

func (s *RaftServer) runElectionTimer() {
	for {
		timeout := randomElectionTimeout()
		time.Sleep(timeout)

		s.mu.Lock()
		// If a heartbeat arrived during our sleep, lastHeartbeat will be recent.
		// Only start an election if no heartbeat has been received within the timeout
		// window and we are still a follower.
		if time.Since(s.lastHeartbeat) < timeout || s.state != raft.Follower {
			s.mu.Unlock()
			continue
		}
		s.state = raft.Candidate
		s.currentTerm++
		s.votedFor = 0
		s.mu.Unlock()
	}
}

func resetElectionTimer(s *RaftServer) {
	s.lastHeartbeat = time.Now()
	s.timerResetCount++
}

func (s *RaftServer) getLastLogIndex() uint64 {
	if len(s.log) == 0 {
		return 0
	}
	return s.log[len(s.log)-1].Index
}

func (s *RaftServer) getLastLogTerm() uint64 {
	if len(s.log) == 0 {
		return 0
	}
	return s.log[len(s.log)-1].Term
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
	//   Tests: commitIndex advances on append and heartbeat, clamped, no regression
	//   TODO: apply committed entries to KV state machine (Phase 2)

	s.mu.Lock()
	defer s.mu.Unlock()

	if req.Term < s.currentTerm {
		return &pb.AppendEntriesResponse{
			Term:    s.currentTerm,
			Success: false,
		}, nil
	}

	if req.Term > s.currentTerm {
		s.currentTerm = req.Term
		s.state = raft.Follower
	} else if req.Term == s.currentTerm && s.state == raft.Candidate {
		s.state = raft.Follower
	}

	if req.PrevLogIndex > 0 {
		if req.PrevLogIndex > uint64(len(s.log)) {
			resetElectionTimer(s)
			return &pb.AppendEntriesResponse{
				Term:    s.currentTerm,
				Success: false,
			}, nil
		}
		if s.log[req.PrevLogIndex-1].Term != req.PrevLogTerm {
			resetElectionTimer(s)
			return &pb.AppendEntriesResponse{
				Term:    s.currentTerm,
				Success: false,
			}, nil
		}
	}

	if len(req.Entries) == 0 {
		// Valid heartbeat: advance commitIndex, reset timer.
		// Leaders piggyback leaderCommit on heartbeats to drive follower state machines.
		if req.LeaderCommit > s.commitIndex {
			s.commitIndex = min(req.LeaderCommit, s.getLastLogIndex())
		}
		resetElectionTimer(s)
		return &pb.AppendEntriesResponse{
			Term:    s.currentTerm,
			Success: true,
		}, nil
	}

	// Append new entries, truncating any conflicts
	insertIndex := req.PrevLogIndex
	for _, entry := range req.Entries {
		if insertIndex < uint64(len(s.log)) {
			if s.log[insertIndex].Term != entry.Term {
				s.log = s.log[:insertIndex] // truncate conflicting entry and all that follow
			} else {
				insertIndex++
				continue // entry already matches, skip to next
			}
		}
		s.log = append(s.log, entry)
		insertIndex++
	}

	if req.LeaderCommit > s.commitIndex {
		s.commitIndex = min(req.LeaderCommit, s.getLastLogIndex())
	}

	resetElectionTimer(s)
	return &pb.AppendEntriesResponse{
		Term:    s.currentTerm,
		Success: true,
	}, nil
}

// NewRaftServer creates and returns a new RaftServer instance.
func NewRaftServer() *RaftServer {
	s := &RaftServer{
		state:         raft.Follower,
		lastHeartbeat: time.Now(),
	}
	go s.runElectionTimer()
	return s
}

// Register wires the RaftServer into a gRPC server.
func (s *RaftServer) Register(grpcServer *grpc.Server) {
	pb.RegisterRaftServiceServer(grpcServer, s)
}
