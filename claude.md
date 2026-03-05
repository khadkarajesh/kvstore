# High-Level Objective
Build a Linearizable, Replicated Key-Value Store using the Raft Consensus Algorithm in Go. The goal is to survive node failures and network partitions while maintaining strict data consistency.

# Tech Stack & Tools
- Language: Go 1.21+ (utilizing Generics and advanced Concurrency)
- RPC: gRPC + Protocol Buffers (protoc)
- Persistence: bbolt (Key-value store for the WAL) or standard os file append.
- Testing: go test -v -race ./... (The -race detector is mandatory for Raft).

# System Architecture
The system consists of $N$ nodes (typically 3 or 5). 
1. Each node runs:Consensus Module: Manages the Raft state machine (Follower/Candidate/Leader).
2. Log Manager: Handles the Write-Ahead Log (WAL) and disk persistence.
3. KV State Machine: The actual in-memory map applied only after consensus.

# Implementation Roadmap
Phase 1: The Heartbeat & Election (Status: 🟡 In Progress)
- [ ] Initialize Go module and project structure.
- [ ] Define raft.proto for gRPC communication.
- [ ] Implement the main Node loop with randomized election timers ($150ms-300ms$).
- [ ] Implement RequestVote RPC (Leader election logic).
- [ ] Implement AppendEntries RPC (Heartbeat logic).

Phase 2: Log Replication (Status: 🔴 Not Started)
- [ ] Implement Log Entry structure.
- [ ] Handle the "Quorum" logic (majority match index).
- [ ] Implement Log consistency checks (term/index matching).
- [ ] Client Request handling (Forwarding from Follower to Leader).

Phase 3: Persistence & Safety (Status: 🔴 Not Started)
- [ ] Implement Disk Persistence for currentTerm and votedFor.
- [ ] Implement Log Compaction (Snapshoting).
- [ ] Failure Testing: Simulate network partitions using iptables or proxy middleware.

# 📂 Project Structure
* `/raft`: Core consensus logic & state machine transitions.
* `/server`: gRPC service implementation and node bootstrapping.
* `/proto`: Protobuf definitions.
* `/pb`: Generated Go code for RPCs.
* `/kv`: The underlying storage engine (State Machine).

# Development Commands
- Generate Proto: protoc --go_out=. --go-grpc_out=. proto/raft.proto
- Run Node: go run main.go -id=1 -port=50051
- Test Suite: go test -v ./raft
- Check Races: go test -race ./...

# Hard Constraints
- No Consensus Libraries: No using etcd/raft or hashicorp/raft.
- Consistency: Must pass the "Split Brain" test (2 nodes cannot both be Leader for the same Term).
- Linearizability: Once a write is acknowledged, all subsequent reads must return that value.