package server

import (
	"context"
	"testing"

	pb "github.com/rajesh/kvstore/proto"
	raft "github.com/rajesh/kvstore/raft"
)

func newServer(term uint64, votedFor uint32, lastLogIndex, lastLogTerm uint64) *RaftServer {
	return &RaftServer{
		state:        raft.Follower,
		currentTerm:  int64(term),
		votedFor:     votedFor,
		lastLogIndex: lastLogIndex,
		lastLogTerm:  lastLogTerm,
	}
}

func TestRequestVote_RejectStaleTerm(t *testing.T) {
	s := newServer(5, 0, 0, 0)

	resp, err := s.RequestVote(context.Background(), &pb.RequestVoteRequest{
		Term:        3, // older than currentTerm=5
		CandidateId: 2,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.VoteGranted {
		t.Error("expected vote denied for stale term, got granted")
	}
	if resp.Term != 5 {
		t.Errorf("expected term=5 in response, got %d", resp.Term)
	}
}

func TestRequestVote_GrantVoteFirstTime(t *testing.T) {
	s := newServer(1, 0, 0, 0) // term=1, hasn't voted yet

	resp, err := s.RequestVote(context.Background(), &pb.RequestVoteRequest{
		Term:        1,
		CandidateId: 2,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.VoteGranted {
		t.Error("expected vote granted, got denied")
	}
	if resp.Term != 1 {
		t.Errorf("expected term=1 in response, got %d", resp.Term)
	}
	if s.votedFor != 2 {
		t.Errorf("expected votedFor=2, got %d", s.votedFor)
	}
}

func TestRequestVote_DenyDoubleVote(t *testing.T) {
	s := newServer(1, 1, 0, 0)
	resp, err := s.RequestVote(context.Background(), &pb.RequestVoteRequest{
		Term:        1,
		CandidateId: 2,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.VoteGranted {
		t.Error("vote granted is not expected")
	}

	if resp.Term != 1 {
		t.Errorf("expected term=1 in response, got %d", resp.Term)
	}
}
