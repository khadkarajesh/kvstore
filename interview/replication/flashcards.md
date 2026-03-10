Replication — Flashcards

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

**Q: What is replication and how does it differ from sharding?**
A: Replication → same data copied to multiple machines (fault tolerance, read scale).
   Sharding → different data on different machines (storage/write scale).
   They compose: each shard can be replicated across multiple nodes.

---

**Q: What are the three purposes of replication?**
A: 1. Fault tolerance — survive node failures without data loss.
   2. Read scalability — distribute reads across replicas.
   3. Reduced latency — serve reads from geographically nearby replica.

---

**Q: What is synchronous replication? What are its tradeoffs?**
A: Leader waits for follower to confirm write before ACK-ing client.
   Pro: no data loss on leader failure — follower is fully current.
   Con: write latency increases; slow or down follower blocks writes.

---

**Q: What is asynchronous replication? What are its tradeoffs?**
A: Leader ACKs client immediately; replication happens in background.
   Pro: low write latency; follower failures don't block writes.
   Con: replication lag; data loss on leader failure if writes not yet replicated.

---

**Q: What is semi-synchronous replication?**
A: Exactly one follower is kept synchronous; all others are async.
   Guarantees at least two nodes (leader + one sync follower) always have the latest data.
   If the sync follower falls behind, an async follower is promoted to sync.

---

**Q: What is the read-your-own-writes (RYOW) problem?**
A: User writes, then reads from a stale follower and sees their write missing.
   Fix: route reads of user-owned data to the leader for a short window after any write,
   or use session tokens to track last-write timestamp and avoid replicas lagging behind it.

---

**Q: What is the monotonic reads problem?**
A: User reads from Replica A (sees V2), then from Replica B (lag is higher, sees V1).
   Time appears to go backwards — a newer value is replaced by an older one across requests.
   Fix: sticky sessions — always route the same user to the same replica.

---

**Q: What is the causal consistency problem in replication?**
A: A user sees a reply before the original post, or an effect before its cause.
   Happens when causally related writes replicate out of order to different followers.
   Fix: causal consistency tracking — replicas apply writes in causal order.

---

**Q: What are the use cases for multi-leader replication?**
A: Multi-datacenter (one leader per DC for local-write low latency), offline clients
   (each device is its own leader, syncs on reconnect), collaborative editing
   (each user/cursor is a leader).

---

**Q: What is the core problem with multi-leader replication?**
A: Write conflicts — two leaders accept different writes for the same key simultaneously.
   Neither write is inherently wrong; the system must resolve which value survives.

---

**Q: What are the three main conflict resolution strategies in multi-leader replication?**
A: Last Write Wins (LWW): highest timestamp survives — simple but risks silent data loss from clock skew.
   CRDTs: data structures that merge automatically by construction (counters, sets).
   Application-level merge: return all conflicting versions to application; app decides (Dynamo).

---

**Q: What is leaderless replication and how does it guarantee consistency?**
A: No single leader. Client writes to W nodes; reads from R nodes.
   W + R > N ensures at least one node in every read quorum has the latest write.
   Conflicts detected via vector clocks; repaired via read repair and anti-entropy.

---

**Q: What is the split brain problem in leader-follower failover?**
A: Leader is partitioned (not dead) — followers elect a new leader. Old leader is still up
   and accepting writes from clients that can reach it. Two nodes both believe they are leader
   → conflicting writes → data corruption.

---

**Q: What are fencing tokens and how do they prevent split brain?**
A: Each leader election issues a monotonically increasing token. Every write must include
   the current token. Storage rejects any write with a token lower than the highest seen.
   Old leader's writes are silently rejected — it cannot corrupt data even if it's still running.

---

**Q: What is STONITH?**
A: Shoot The Other Node In The Head. New leader sends a kill/power-off signal to the old leader
   via out-of-band network (IPMI/iDRAC) before accepting writes. Forcibly prevents old leader
   from accepting any more writes. Used in high-stakes DB failover.

---

**Q: How does Raft prevent split brain without STONITH or fencing tokens?**
A: Epoch/term fencing. Every RPC includes the sender's current term.
   If any node receives an RPC with a higher term, it immediately steps down to follower.
   Old leader self-demotes the moment it receives any message from the new leader's term.

---

**Q: What is the data loss risk during failover in async replication?**
A: New leader is elected from the most up-to-date follower. But with async replication,
   that follower may still be behind. Writes acknowledged by the old leader but not yet
   replicated are permanently lost. Mitigation: elect follower with highest replication offset.

---

**Q: Senior interview script for replication design?**
A: "Async leader-follower for read scalability. Writes to leader; reads to followers.
   RYOW handled by routing post-write reads to leader within a 30-second window.
   Follower lag monitored — remove from read pool if lag exceeds threshold.
   Failover uses fencing tokens so old leader's writes are rejected at storage layer."
