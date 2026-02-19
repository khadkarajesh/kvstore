package server

import (
	"context"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb   "github.com/rajesh/kvstore/proto"
	raft "github.com/rajesh/kvstore/raft"
)

// RaftServer implements the gRPC RaftServiceServer interface.
type RaftServer struct {
	pb.UnimplementedRaftServiceServer
	mu sync.Mutex
	state raft.NodeState
	currentTerm int64
	votedFor    uint32 
    lastLogIndex uint64
    lastLogTerm  uint64
}

func (s *RaftServer) RequestVote(ctx context.Context, req *pb.RequestVoteRequest) (*pb.RequestVoteResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if req.Term < uint64(s.currentTerm) {
		return &pb.RequestVoteResponse{
			Term: uint64(s.currentTerm),
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

func (s *RaftServer) AppendEntries(ctx context.Context, req *pb.AppendEntriesRequest) (*pb.AppendEntriesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "AppendEntries not implemented")
}

// NewRaftServer creates and returns a new RaftServer instance.
func NewRaftServer() *RaftServer {
	return &RaftServer{}
}

// Register wires the RaftServer into a gRPC server.
func (s *RaftServer) Register(grpcServer *grpc.Server) {
	pb.RegisterRaftServiceServer(grpcServer, s)
}
