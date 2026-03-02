package server

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	pb "github.com/rajesh/kvstore/proto"
	raft "github.com/rajesh/kvstore/raft"
)

var errSteppedDown = errors.New("node stepped down")

type applyResult struct {
	value string
	found bool
	err   error
}

type VoteReply struct {
	Term        uint64
	VoteGranted bool
}

type Peer struct {
	ID       uint32
	Addr     string
	Client   pb.RaftServiceClient
	KVClient pb.KVServiceClient
	Conn     *grpc.ClientConn
}

// RaftServer implements the gRPC RaftServiceServer and KVServiceServer interfaces.
type RaftServer struct {
	pb.UnimplementedRaftServiceServer
	pb.UnimplementedKVServiceServer

	id              uint32
	mu              sync.Mutex
	state           raft.NodeState
	currentTerm     uint64
	votedFor        uint32
	leaderID        uint32
	lastHeartbeat   time.Time
	timerResetCount int
	log             []*pb.LogEntry
	commitIndex     uint64

	// Leader-only replication state (initialized in becomeLeader).
	nextIndex  map[uint32]uint64
	matchIndex map[uint32]uint64
	triggerCh  map[uint32]chan struct{}

	// KV state machine (all nodes).
	kv          map[string]string
	lastApplied uint64
	pending     map[uint64]chan applyResult
	applyCond   *sync.Cond

	peers map[uint32]*Peer
}

// ── Election ──────────────────────────────────────────────────────────────────

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

func (s *RaftServer) sendRequestVote(peer *Peer) *VoteReply {
	s.mu.Lock()
	term := s.currentTerm
	id := s.id
	lastLogIndex := s.getLastLogIndex()
	lastLogTerm := s.getLastLogTerm()
	s.mu.Unlock()

	resp, err := peer.Client.RequestVote(context.Background(), &pb.RequestVoteRequest{
		Term:         term,
		CandidateId:  id,
		LastLogIndex: lastLogIndex,
		LastLogTerm:  lastLogTerm,
	})
	if err != nil || resp == nil {
		return &VoteReply{}
	}
	return &VoteReply{
		Term:        resp.Term,
		VoteGranted: resp.VoteGranted,
	}
}

func (s *RaftServer) runElectionTimer() {
	for {
		timeout := randomElectionTimeout()
		time.Sleep(timeout)

		s.mu.Lock()
		if time.Since(s.lastHeartbeat) < timeout || s.state != raft.Follower {
			s.mu.Unlock()
			continue
		}
		voteCh := make(chan *VoteReply, len(s.peers))
		s.state = raft.Candidate
		s.currentTerm++
		s.votedFor = s.id
		votes := 1

		for _, peer := range s.peers {
			p := peer
			go func(pp *Peer) {
				voteCh <- s.sendRequestVote(pp)
			}(p)
		}
		s.mu.Unlock()

		deadline := time.After(timeout)
		remaining := len(s.peers)
		stepDown := false

		for remaining > 0 {
			select {
			case voteReply := <-voteCh:
				s.mu.Lock()
				if voteReply.Term > s.currentTerm {
					s.state = raft.Follower
					votes = 0
					stepDown = true
					s.currentTerm = voteReply.Term
					s.votedFor = 0
				}
				if voteReply.VoteGranted && voteReply.Term == s.currentTerm {
					votes++
				}
				remaining--
				s.mu.Unlock()
			case <-deadline:
				remaining = 0
			}
			if stepDown || votes > (len(s.peers)+1)/2 {
				break
			}
		}
		s.mu.Lock()
		if s.state == raft.Candidate {
			if votes > (len(s.peers)+1)/2 {
				s.becomeLeader()
			} else {
				s.state = raft.Follower
			}
		}
		s.mu.Unlock()
	}
}

// ── Leader ────────────────────────────────────────────────────────────────────

func (s *RaftServer) becomeLeader() {
	s.state = raft.Leader
	s.leaderID = s.id
	lastIdx := s.getLastLogIndex()
	s.nextIndex = make(map[uint32]uint64)
	s.matchIndex = make(map[uint32]uint64)
	s.triggerCh = make(map[uint32]chan struct{})
	for id, peer := range s.peers {
		s.nextIndex[id] = lastIdx + 1
		s.matchIndex[id] = 0
		ch := make(chan struct{}, 1)
		s.triggerCh[id] = ch
		go s.replicateToPeer(peer, s.currentTerm, ch)
	}
}

// replicateToPeer sends AppendEntries to one peer, handling both heartbeats
// (every 50 ms) and new entries (triggered via ch). Bound to term; exits when
// the node is no longer the leader for that term.
func (s *RaftServer) replicateToPeer(peer *Peer, term uint64, ch chan struct{}) {
	for {
		select {
		case <-ch:
		case <-time.After(50 * time.Millisecond):
		}

		s.mu.Lock()
		if s.state != raft.Leader || s.currentTerm != term {
			s.mu.Unlock()
			return
		}

		nextIdx := s.nextIndex[peer.ID]
		prevLogIndex := nextIdx - 1
		prevLogTerm := uint64(0)
		if prevLogIndex > 0 {
			prevLogTerm = s.log[prevLogIndex-1].Term
		}

		startIdx := int(nextIdx - 1)
		var entries []*pb.LogEntry
		if startIdx < len(s.log) {
			src := s.log[startIdx:]
			entries = make([]*pb.LogEntry, len(src))
			copy(entries, src)
		}

		currentTerm := s.currentTerm
		leaderCommit := s.commitIndex
		id := s.id
		s.mu.Unlock()

		resp, err := peer.Client.AppendEntries(context.Background(), &pb.AppendEntriesRequest{
			Term:         currentTerm,
			LeaderId:     id,
			PrevLogIndex: prevLogIndex,
			PrevLogTerm:  prevLogTerm,
			Entries:      entries,
			LeaderCommit: leaderCommit,
		})

		if err != nil || resp == nil {
			continue
		}

		s.mu.Lock()
		if resp.Term > s.currentTerm {
			s.clearPending(errSteppedDown)
			s.state = raft.Follower
			s.currentTerm = resp.Term
			s.votedFor = 0
			s.mu.Unlock()
			return
		}
		if s.state != raft.Leader || s.currentTerm != term {
			s.mu.Unlock()
			return
		}
		if resp.Success {
			newMatch := prevLogIndex + uint64(len(entries))
			if newMatch > s.matchIndex[peer.ID] {
				s.matchIndex[peer.ID] = newMatch
				s.nextIndex[peer.ID] = newMatch + 1
			}
			s.maybeAdvanceCommitIndex()
		} else {
			if s.nextIndex[peer.ID] > 1 {
				s.nextIndex[peer.ID]--
			}
			// Schedule immediate retry.
			select {
			case ch <- struct{}{}:
			default:
			}
		}
		s.mu.Unlock()
	}
}

// maybeAdvanceCommitIndex checks whether a new commitIndex can be established
// based on matchIndex across peers. Must be called with s.mu held.
func (s *RaftServer) maybeAdvanceCommitIndex() {
	lastIdx := s.getLastLogIndex()
	for n := lastIdx; n > s.commitIndex; n-- {
		// Safety: only commit entries from the current term (Raft §5.4.2).
		if s.log[n-1].Term != s.currentTerm {
			continue
		}
		count := 1 // self
		for _, peer := range s.peers {
			if s.matchIndex[peer.ID] >= n {
				count++
			}
		}
		if count > (len(s.peers)+1)/2 {
			s.commitIndex = n
			s.applyCond.Signal()
			break
		}
	}
}

// triggerReplication does a non-blocking send to every peer's trigger channel.
// Must be called with s.mu held.
func (s *RaftServer) triggerReplication() {
	for _, ch := range s.triggerCh {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// clearPending signals all waiting Put/Get handlers that the leader stepped
// down. Must be called with s.mu held.
func (s *RaftServer) clearPending(err error) {
	for idx, ch := range s.pending {
		ch <- applyResult{err: err}
		delete(s.pending, idx)
	}
}

// ── Apply loop ────────────────────────────────────────────────────────────────

func (s *RaftServer) applyLoop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for {
		for s.lastApplied >= s.commitIndex {
			s.applyCond.Wait()
		}
		for s.lastApplied < s.commitIndex {
			s.lastApplied++
			entry := s.log[s.lastApplied-1]

			var cmd pb.KVCommand
			if err := proto.Unmarshal(entry.Command, &cmd); err != nil {
				continue
			}

			var result applyResult
			switch cmd.Op {
			case "put":
				s.kv[cmd.Key] = cmd.Value
			case "get":
				val, found := s.kv[cmd.Key]
				result = applyResult{value: val, found: found}
			}

			if ch, ok := s.pending[entry.Index]; ok {
				ch <- result
				delete(s.pending, entry.Index)
			}
		}
	}
}

// ── Log helpers ───────────────────────────────────────────────────────────────

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

// ── AppendEntries ─────────────────────────────────────────────────────────────

func (s *RaftServer) AppendEntries(ctx context.Context, req *pb.AppendEntriesRequest) (*pb.AppendEntriesResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if req.Term < s.currentTerm {
		return &pb.AppendEntriesResponse{
			Term:    s.currentTerm,
			Success: false,
		}, nil
	}

	if req.Term > s.currentTerm {
		if s.state == raft.Leader {
			s.clearPending(errSteppedDown)
		}
		s.currentTerm = req.Term
		s.state = raft.Follower
		s.votedFor = 0
	} else if req.Term == s.currentTerm && s.state == raft.Candidate {
		s.state = raft.Follower
	}

	s.leaderID = req.LeaderId

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
		if req.LeaderCommit > s.commitIndex {
			s.commitIndex = min(req.LeaderCommit, s.getLastLogIndex())
			s.applyCond.Signal()
		}
		resetElectionTimer(s)
		return &pb.AppendEntriesResponse{
			Term:    s.currentTerm,
			Success: true,
		}, nil
	}

	insertIndex := req.PrevLogIndex
	for _, entry := range req.Entries {
		if insertIndex < uint64(len(s.log)) {
			if s.log[insertIndex].Term != entry.Term {
				s.log = s.log[:insertIndex]
			} else {
				insertIndex++
				continue
			}
		}
		s.log = append(s.log, entry)
		insertIndex++
	}

	if req.LeaderCommit > s.commitIndex {
		s.commitIndex = min(req.LeaderCommit, s.getLastLogIndex())
		s.applyCond.Signal()
	}

	resetElectionTimer(s)
	return &pb.AppendEntriesResponse{
		Term:    s.currentTerm,
		Success: true,
	}, nil
}

// ── KV handlers ───────────────────────────────────────────────────────────────

func (s *RaftServer) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	s.mu.Lock()
	if s.state != raft.Leader {
		peer := s.peers[s.leaderID]
		s.mu.Unlock()
		if peer == nil {
			return &pb.PutResponse{Success: false}, nil
		}
		return peer.KVClient.Put(ctx, req)
	}

	cmdBytes, err := proto.Marshal(&pb.KVCommand{Op: "put", Key: req.Key, Value: req.Value})
	if err != nil {
		s.mu.Unlock()
		return &pb.PutResponse{Success: false}, err
	}
	index := s.getLastLogIndex() + 1
	s.log = append(s.log, &pb.LogEntry{Term: s.currentTerm, Index: index, Command: cmdBytes})
	doneCh := make(chan applyResult, 1)
	s.pending[index] = doneCh
	s.maybeAdvanceCommitIndex() // handles single-node cluster
	s.triggerReplication()
	s.mu.Unlock()

	select {
	case r := <-doneCh:
		if r.err != nil {
			return &pb.PutResponse{Success: false}, r.err
		}
		return &pb.PutResponse{Success: true}, nil
	case <-ctx.Done():
		s.mu.Lock()
		delete(s.pending, index)
		s.mu.Unlock()
		return &pb.PutResponse{Success: false}, ctx.Err()
	}
}

func (s *RaftServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	s.mu.Lock()
	if s.state != raft.Leader {
		peer := s.peers[s.leaderID]
		s.mu.Unlock()
		if peer == nil {
			return &pb.GetResponse{}, nil
		}
		return peer.KVClient.Get(ctx, req)
	}

	cmdBytes, err := proto.Marshal(&pb.KVCommand{Op: "get", Key: req.Key})
	if err != nil {
		s.mu.Unlock()
		return &pb.GetResponse{}, err
	}
	index := s.getLastLogIndex() + 1
	s.log = append(s.log, &pb.LogEntry{Term: s.currentTerm, Index: index, Command: cmdBytes})
	doneCh := make(chan applyResult, 1)
	s.pending[index] = doneCh
	s.maybeAdvanceCommitIndex() // handles single-node cluster
	s.triggerReplication()
	s.mu.Unlock()

	select {
	case r := <-doneCh:
		if r.err != nil {
			return &pb.GetResponse{}, r.err
		}
		return &pb.GetResponse{Value: r.value, Found: r.found}, nil
	case <-ctx.Done():
		s.mu.Lock()
		delete(s.pending, index)
		s.mu.Unlock()
		return &pb.GetResponse{}, ctx.Err()
	}
}

// ── Bootstrap ─────────────────────────────────────────────────────────────────

// NewRaftServer creates and returns a new RaftServer instance.
// peers must be fully dialed before calling — the election timer starts immediately.
func NewRaftServer(id uint32, peers map[uint32]*Peer) *RaftServer {
	s := &RaftServer{
		id:            id,
		peers:         peers,
		state:         raft.Follower,
		lastHeartbeat: time.Now(),
		kv:            make(map[string]string),
		pending:       make(map[uint64]chan applyResult),
	}
	s.applyCond = sync.NewCond(&s.mu)
	go s.runElectionTimer()
	go s.applyLoop()
	return s
}

// Register wires the RaftServer into a gRPC server.
func (s *RaftServer) Register(grpcServer *grpc.Server) {
	pb.RegisterRaftServiceServer(grpcServer, s)
	pb.RegisterKVServiceServer(grpcServer, s)
}
