package server

import (
	"context"
	"testing"

	pb "github.com/rajesh/kvstore/proto"
	raft "github.com/rajesh/kvstore/raft"
)

func newServer(term uint64, votedFor uint32, lastLogIndex, lastLogTerm uint64) *RaftServer {
	s := &RaftServer{
		state:       raft.Follower,
		currentTerm: term,
		votedFor:    votedFor,
	}
	if lastLogIndex > 0 {
		s.log = append(s.log, &pb.LogEntry{Index: lastLogIndex, Term: lastLogTerm})
	}
	return s
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
		t.Errorf("expected currentTerm=2, got %d", s.currentTerm)
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

func TestAppendEntries_RejectStaleTerm(t *testing.T) {
	s := newServer(5, 0, 0, 0)
	prevCount := s.timerResetCount

	resp, err := s.AppendEntries(context.Background(), &pb.AppendEntriesRequest{
		Term: 3, // older than currentTerm=5
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Error("expected AppendEntries rejected for stale term, got success")
	}

	if resp.Term != 5 {
		t.Errorf("expected resp.Term=5, got %d", resp.Term)
	}

	if s.timerResetCount != prevCount {
		t.Error("expected election timer NOT to reset on stale heartbeat, but it did")
	}
}

func TestAppendEntries_AcceptHeartbeat(t *testing.T) {
	s := newServer(5, 0, 0, 0)
	prevCount := s.timerResetCount
	resp, err := s.AppendEntries(context.Background(), &pb.AppendEntriesRequest{
		Term: 5, // matches currentTerm
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Error("expected AppendEntries accepted for valid heartbeat, got rejected")
	}

	if resp.Term != 5 {
		t.Errorf("expected resp.Term=5, got %d", resp.Term)
	}

	if s.timerResetCount == prevCount {
		t.Error("expected election timer to reset on valid heartbeat, but timerResetCount did not change")
	}
}

func TestAppendEntries_StepDownOnValidHeartbeat(t *testing.T) {
	tests := []struct {
		name      string
		initState raft.NodeState
		initTerm  uint64
		reqTerm   uint64
		wantTerm  uint64
	}{
		{"follower updates term on higher term", raft.Follower, 5, 7, 7},
		{"candidate steps down on same term", raft.Candidate, 2, 2, 2},
		{"candidate steps down on higher term", raft.Candidate, 2, 5, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newServer(tt.initTerm, 0, 0, 0)
			s.state = tt.initState
			prevCount := s.timerResetCount

			resp, err := s.AppendEntries(context.Background(), &pb.AppendEntriesRequest{
				Term: tt.reqTerm,
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp.Term != tt.wantTerm {
				t.Errorf("expected wantTerm=%d, got %d", tt.wantTerm, resp.Term)
			}

			if s.currentTerm != tt.wantTerm {
				t.Errorf("expected currentTerm=%d, got %d", tt.wantTerm, s.currentTerm)
			}

			if s.state != raft.Follower {
				t.Errorf("expected state to be Follower, got %v", s.state)
			}

			if s.timerResetCount == prevCount {
				t.Error("expected election timer to reset on valid heartbeat, but timerResetCount did not change")
			}

			if !resp.Success {
				t.Error("expected Success=true on valid heartbeat, got false")
			}

		})
	}
}

func TestAppendEntries_AppendsFirstEntry(t *testing.T) {
	s := newServer(0, 0, 0, 0)
	resp, err := s.AppendEntries(context.Background(), &pb.AppendEntriesRequest{
		Term: 1,
		Entries: []*pb.LogEntry{
			{Index: 1, Term: 1, Command: []byte("set x=1")},
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !resp.Success {
		t.Error("expected AppendEntries accepted for first entry, got rejected")
	}

	if len(s.log) != 1 {
		t.Errorf("expected log length=1, got %d", len(s.log))
	}

}

func TestAppendEntries_MatchingPreviousEntries(t *testing.T) {
	s := newServer(1, 0, 0, 0)
	s.log = []*pb.LogEntry{
		{Index: 1, Term: 1},
		{Index: 2, Term: 2},
	}
	req := &pb.AppendEntriesRequest{
		Term:         2,
		PrevLogIndex: 2,
		PrevLogTerm:  2,
		Entries: []*pb.LogEntry{
			{Index: 3, Term: 2, Command: []byte("set x=1")},
		},
	}
	resp, err := s.AppendEntries(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !resp.Success {
		t.Error("expected AppendEntries accepted for matching previous entry, got rejected")
	}

	if len(s.log) != 3 {
		t.Errorf("expected log length=3, got %d", len(s.log))
	}

	if s.log[2].Command == nil || string(s.log[2].Command) != "set x=1" {
		t.Errorf("expected log[2].Command='set x=1', got %s", string(s.log[2].Command))
	}

	if req.PrevLogIndex != s.log[1].Index || req.PrevLogTerm != s.log[1].Term {
		t.Errorf("expected PrevLogIndex=%d and PrevLogTerm=%d, got PrevLogIndex=%d and PrevLogTerm=%d",
			s.log[1].Index, s.log[1].Term, req.PrevLogIndex, req.PrevLogTerm)
	}
}

func TestAppendEntries_GapInPreviousEntries(t *testing.T) {
	s := newServer(1, 0, 0, 0)
	s.log = []*pb.LogEntry{
		{Index: 1, Term: 1},
		{Index: 2, Term: 2},
	}
	req := &pb.AppendEntriesRequest{
		Term:         2,
		PrevLogIndex: 3, // gap after index 2
		PrevLogTerm:  2,
		Entries: []*pb.LogEntry{
			{Index: 4, Term: 2, Command: []byte("set x=1")},
		},
	}
	resp, err := s.AppendEntries(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Success {
		t.Error("expected AppendEntries rejected for gap in previous entries, got accepted")
	}

	if len(s.log) != 2 {
		t.Errorf("expected log length=2 after rejection, got %d", len(s.log))
	}

	if resp.Term != 2 {
		t.Errorf("expected resp.Term=2, got %d", resp.Term)
	}

	if s.timerResetCount == 0 {
		t.Error("expected election timer to reset on gap rejection, but timerResetCount did not change")
	}
}

func TestAppendEntries_ConflictInPreviousEntries(t *testing.T) {
	s := newServer(1, 0, 0, 0)
	s.log = []*pb.LogEntry{
		{Index: 1, Term: 1},
		{Index: 2, Term: 2},
	}
	req := &pb.AppendEntriesRequest{
		Term:         2,
		PrevLogIndex: 2,
		PrevLogTerm:  1, // conflict with log[1].Term=2
		Entries: []*pb.LogEntry{
			{Index: 3, Term: 2, Command: []byte("set x=1")},
		},
	}
	resp, err := s.AppendEntries(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Success {
		t.Error("expected AppendEntries rejected for conflict in previous entries, got accepted")
	}

	if len(s.log) != 2 {
		t.Errorf("expected log length=2, got %d", len(s.log))
	}

	if resp.Term != 2 {
		t.Errorf("expected resp.Term=2, got %d", resp.Term)
	}

	if s.timerResetCount == 0 {
		t.Error("expected election timer to reset on AppendEntries rejection due to log conflict, but timerResetCount did not change")
	}

}

func TestAppendEntries_HeartbeatWithPrevLogIndexBeyondLog(t *testing.T) {
	s := newServer(1, 0, 0, 0)
	s.log = []*pb.LogEntry{
		{Index: 1, Term: 1},
		{Index: 2, Term: 2},
	}
	prevCount := s.timerResetCount

	// Heartbeat (no entries) where leader believes follower is at index 5,
	// but follower only has 2 entries. Consistency check must still run.
	resp, err := s.AppendEntries(context.Background(), &pb.AppendEntriesRequest{
		Term:         2,
		PrevLogIndex: 5,
		PrevLogTerm:  2,
		Entries:      nil,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Success {
		t.Error("expected AppendEntries rejected: heartbeat PrevLogIndex beyond follower log, got accepted")
	}

	if resp.Term != 2 {
		t.Errorf("expected resp.Term=2, got %d", resp.Term)
	}

	if len(s.log) != 2 {
		t.Errorf("expected log length=2, got %d", len(s.log))
	}

	if s.timerResetCount == prevCount {
		t.Error("expected election timer to reset on valid-term heartbeat, but timerResetCount did not change")
	}
}

func TestAppendEntries_CommitIndexAdvancesOnAppend(t *testing.T) {
	s := newServer(1, 0, 0, 0)
	s.log = []*pb.LogEntry{
		{Index: 1, Term: 1},
		{Index: 2, Term: 1},
	}

	_, err := s.AppendEntries(context.Background(), &pb.AppendEntriesRequest{
		Term:         1,
		PrevLogIndex: 2,
		PrevLogTerm:  1,
		Entries:      []*pb.LogEntry{{Index: 3, Term: 1, Command: []byte("set x=1")}},
		LeaderCommit: 2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// leaderCommit=2, lastLogIndex=3 → commitIndex = min(2,3) = 2
	if s.commitIndex != 2 {
		t.Errorf("expected commitIndex=2, got %d", s.commitIndex)
	}
}

func TestAppendEntries_CommitIndexClampedToLastLogIndex(t *testing.T) {
	s := newServer(1, 0, 0, 0)

	_, err := s.AppendEntries(context.Background(), &pb.AppendEntriesRequest{
		Term:         1,
		Entries:      []*pb.LogEntry{{Index: 1, Term: 1, Command: []byte("set x=1")}},
		LeaderCommit: 10, // leader is far ahead; follower only appended index 1
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// leaderCommit=10, lastLogIndex=1 → commitIndex = min(10,1) = 1
	if s.commitIndex != 1 {
		t.Errorf("expected commitIndex=1 (clamped to lastLogIndex), got %d", s.commitIndex)
	}
}

func TestAppendEntries_CommitIndexAdvancesOnHeartbeat(t *testing.T) {
	s := newServer(1, 0, 0, 0)
	s.log = []*pb.LogEntry{
		{Index: 1, Term: 1},
		{Index: 2, Term: 1},
		{Index: 3, Term: 1},
	}
	s.commitIndex = 1

	_, err := s.AppendEntries(context.Background(), &pb.AppendEntriesRequest{
		Term:         1,
		LeaderCommit: 3, // heartbeat: no entries, but leader has committed up to 3
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.commitIndex != 3 {
		t.Errorf("expected commitIndex=3, got %d", s.commitIndex)
	}
}

func TestAppendEntries_CommitIndexDoesNotRegress(t *testing.T) {
	s := newServer(1, 0, 0, 0)
	s.log = []*pb.LogEntry{
		{Index: 1, Term: 1},
		{Index: 2, Term: 1},
		{Index: 3, Term: 1},
	}
	s.commitIndex = 3

	_, err := s.AppendEntries(context.Background(), &pb.AppendEntriesRequest{
		Term:         1,
		LeaderCommit: 2, // stale leaderCommit below current commitIndex
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.commitIndex != 3 {
		t.Errorf("expected commitIndex to stay at 3, got %d", s.commitIndex)
	}
}
