package raft

type NodeState int

const (
	Follower  NodeState = iota // default on startup, waits for heartbeats
	Candidate                  // started an election, waiting for votes
	Leader                     // won the election, sends heartbeats
)

func (s NodeState) String() string {
	return [...]string{"Follower", "Candidate", "Leader"}[s]
}
