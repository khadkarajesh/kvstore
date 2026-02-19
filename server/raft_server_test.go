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

func TestRequestVote_UpdateTermOnNewerTerm(t *testing.T) {
	s := newServer(1, 1, 3, 1)
	_, err := s.RequestVote(context.Background(), &pb.RequestVoteRequest{
		Term:         2, // newer term
		CandidateId:  2,
		LastLogIndex: 0,
		LastLogTerm:  0,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.votedFor != 0 {
		t.Errorf("expected votedFor=0 after term update, got %d", s.votedFor)
	}

	if s.currentTerm != 2 {
		t.Errorf("expected term=2 in response, got %d", s.currentTerm)
	}

	if s.state != raft.Follower {
		t.Errorf("expected state to be Follower, got %v", s.state)
	}
}

func TestRequestVote_StaleLogDenied_OlderTerm(t *testing.T) {
	s := newServer(1, 0, 3, 1)
	req := pb.RequestVoteRequest{
		Term:         1,
		CandidateId:  2,
		LastLogIndex: 0,
		LastLogTerm:  0,
	}
	resp, err := s.RequestVote(context.Background(), &req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// vote denied means votedFor must NOT have changed
	if s.votedFor != 0 {
		t.Errorf("expected votedFor=0, vote should not be recorded on denial, got %d", s.votedFor)
	}

	// response term should reflect current term
	if resp.Term != 1 {
		t.Errorf("expected resp.Term=1, got %d", resp.Term)
	}

	if resp.VoteGranted {
		t.Error("expected vote denied for stale log, got granted")
	}
}

func TestRequestVote_StaleLogDenied_ShorterIndex(t *testing.T) {
	s1 := newServer(1, 0, 3, 1)
	req1 := pb.RequestVoteRequest{
		Term:         1,
		CandidateId:  2,
		LastLogIndex: 0,
		LastLogTerm:  1,
	}

	resp1, err1 := s1.RequestVote(context.Background(), &req1)

	if err1 != nil {
		t.Fatalf("unexpected error: %v", err1)
	}

	if resp1.Term != 1 {
		t.Errorf("expected resp.Term=1, got %d", resp1.Term)
	}

	if s1.votedFor != 0 {
		t.Errorf("expected votedFor=0, vote should not be recorded on denial, got %d", s1.votedFor)
	}

	if resp1.VoteGranted {
		t.Error("expected vote denied for stale log, got granted")
	}
}

func TestRequestVote_IdempotentReVote(t *testing.T) {
	s := newServer(1, 2, 3, 1)
	req := pb.RequestVoteRequest{
		Term:         1,
		CandidateId:  2,
		LastLogIndex: 3,
		LastLogTerm:  1,
	}

	resp, err := s.RequestVote(context.Background(), &req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !resp.VoteGranted {
		t.Error("expected vote granted on idempotent request, got denied")
	}

	if s.votedFor != 2 {
		t.Errorf("expected votedFor=2, got %d", s.votedFor)
	}
}
