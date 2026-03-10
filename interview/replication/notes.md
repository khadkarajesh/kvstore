Replication — FAANG Interview Reference

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. What Replication Is
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Multiple copies of the same data on multiple machines.

  Purposes:
    → fault tolerance: if one machine dies, others serve the data
    → read scalability: distribute reads across replicas
    → reduced read latency: serve from geographically nearby replica

---
Replication vs Sharding — distinct concepts

  Replication → same data on multiple machines (copies)
  Sharding    → different data on different machines (partitions)

  They compose: each shard can itself be replicated.
  Example: 4 shards × 3 replicas per shard = 12 nodes total

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
2. Leader-Follower (Single-Leader) Replication
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Topology:
  One leader → accepts all writes
  N followers → replicate from leader, serve reads

  All writes go through leader → no write conflicts by design

---
Synchronous Replication

  Leader waits for at least one follower to confirm write before ACK-ing client.

  Pros:
    → no data loss on leader failure — follower is fully up to date
    → strong durability guarantee

  Cons:
    → write latency increases (must wait for network round trip to follower)
    → if follower is slow or down, writes block (or time out)

---
Asynchronous Replication

  Leader ACKs client immediately. Replication happens in background.

  Pros:
    → low write latency — no waiting for followers
    → follower failures don't block writes

  Cons:
    → replication lag — follower may be seconds or minutes behind
    → data loss on leader failure if recent writes not yet replicated
    → reads from follower may return stale data

---
Semi-Synchronous Replication

  Exactly one follower is synchronous; the rest are async.
  If the sync follower falls behind → promote an async follower to sync.
  Guarantees: at least two nodes (leader + one follower) always have up-to-date data.

  Balance: durability without blocking on all followers.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
3. Replication Lag Problems
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Read-Your-Own-Writes (RYOW)

  Scenario:
    1. User writes (POST /profile) → goes to leader
    2. User reads (GET /profile) → goes to stale follower
    3. User sees their write has disappeared

  Fixes:
    → route reads of user-owned data to leader (for N seconds after write)
    → track last write timestamp; route to leader if replica lag > timestamp
    → sticky read: after any write, read from leader for ~1 minute

---
Monotonic Reads

  Scenario:
    1. User reads from Replica A (lag = 0s) → sees value V2
    2. User reads from Replica B (lag = 10s) → sees value V1 (older)
    3. Time appears to go backwards from user's perspective

  Fix:
    → sticky sessions: always route same user to same replica
    → track read timestamp: never serve read from replica behind user's last-seen version

---
Summary of Lag Anomalies

  RYOW           → user doesn't see own write
  Monotonic reads → user sees newer value then older value across requests
  Causal          → user sees effect before cause (reply before original post)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
4. Multi-Leader Replication
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Multiple leaders accept writes. Leaders replicate to each other.

  Use cases:
    → multi-datacenter: one leader per datacenter, low-latency local writes
    → offline clients: each device is a leader (mobile app, Google Docs offline)
    → collaborative editing: each user's cursor is a leader

  Problem: write conflicts
    Two leaders accept different writes for same key simultaneously.
    Neither write is wrong — both are legitimate.

---
Conflict Resolution Strategies

  Last Write Wins (LWW):
    → attach timestamp to each write; highest timestamp wins
    → simple, widely used (Cassandra default)
    → risk: clock skew → wrong value survives; data loss is silent

  CRDTs (Conflict-free Replicated Data Types):
    → data structures designed to merge automatically (counters, sets, maps)
    → no conflict possible by construction
    → works for specific use cases (shopping cart, view counts)

  Application-level merge:
    → return all conflicting versions to application
    → application decides how to merge (Dynamo's approach)
    → most flexible, most complex to implement

  Avoid conflicts:
    → route all writes for a given record to the same leader
    → only conflict if that leader fails and traffic reroutes

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
5. Leaderless Replication
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

No leader. Client (or coordinator) writes to multiple nodes. Reads from multiple nodes.

  This is what Dynamo uses.

  Consistency guarantee: quorum
    W + R > N → at least one node in every read quorum has the latest write

  Conflict detection: vector clocks (version vectors)
    Each write tagged with version; reads compare versions to detect concurrent writes

  Repair mechanisms:
    → read repair: on read, detect stale node, write latest version back
    → anti-entropy: background process compares Merkle trees, syncs diverged data

  Key difference from leader-follower:
    → no SPOF on writes (any node accepts)
    → no automatic failover needed
    → consistency is probabilistic, not guaranteed (unless W + R > N)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
6. Failover
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Leader fails → elect new leader (from most up-to-date follower).

  Automatic failover steps:
    1. Detect leader failure (timeout, heartbeat loss)
    2. Elect new leader (Raft, Paxos, or async vote among followers)
    3. Reconfigure clients/load balancer to send writes to new leader
    4. Old leader comes back → must accept follower role

---
Split Brain Risk

  Scenario: leader is partitioned (not dead, just unreachable from followers).
  Followers elect a new leader.
  Old leader is still up and accepting writes from clients that can still reach it.
  Two nodes both think they are the leader → conflicting writes → data corruption.

  Fixes:

  Fencing Tokens:
    → every leader election issues a monotonically increasing token
    → all writes must include current token
    → storage/downstream services reject writes with a lower token than seen before
    → old leader's writes are rejected automatically when its token is stale

  STONITH (Shoot The Other Node In The Head):
    → new leader sends a kill signal (or power-off command) to old leader
    → forcibly terminates old leader before it can accept writes
    → requires out-of-band management network or IPMI access

  Epoch fencing (Raft/Zab approach):
    → leader includes its term number in every RPC
    → if any node sees a higher term, it immediately steps down
    → old leader self-demotes on receiving a higher-term message

---
Other Failover Risks

  New leader may not have latest data (async replication lag):
    → promotes a stale follower → some writes lost
    → mitigation: elect follower with highest replication offset

  Replication position reset:
    → new leader was behind → its log position counter is lower than clients expect
    → old leader returns and confuses clients about current state

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
7. What to Say in a Senior Interview
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  "I would use async leader-follower replication for read scalability.
   Writes go to leader. Reads go to followers for most queries.
   For read-your-own-writes I route the immediate post-write read to the leader,
   or use a 30-second leader-read window keyed on session token.
   Follower lag is monitored; if lag exceeds threshold, that follower is taken out
   of the read pool. Failover uses fencing tokens to prevent split-brain —
   the new leader's token is higher, old leader's writes are rejected at storage layer."

---
Decision Tree

  Need write scalability across datacenters?   → multi-leader (one leader/DC)
  Need write scalability, conflict-tolerant?   → leaderless (Dynamo/Cassandra)
  Single region, strong consistency required?  → single-leader, sync or semi-sync
  Single region, read scalability, ok with lag → single-leader, async
