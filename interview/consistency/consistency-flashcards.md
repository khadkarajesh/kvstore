## Consistency Models — Anki Flashcards

Format: Q → A
Cover the answer, recall it, then check.

---

### Core Concepts

**Q: What is the core question that separates all consistency models?**
A: After a write happens — who is guaranteed to see it, and when?

---

**Q: What are the four consistency models in order from weakest to strongest?**
A: Eventual → Session (read your own writes) → Causal → Strong (Linearizable)

---

**Q: What does each step up the consistency spectrum cost?**
A: More coordination between nodes → higher latency → lower availability.

---

**Q: What is the CAP theorem real-world choice?**
A: Partitions always happen at scale. Real choice is C vs A — consistency or availability during a partition.

---

### Eventual Consistency

**Q: What does eventual consistency guarantee?**
A: All replicas converge eventually. No timing promise. Stale reads are possible for anyone.

---

**Q: What is eventual consistency's CAP choice?**
A: AP — available during partition, accepts stale reads.

---

**Q: Name four real-world systems that use eventual consistency.**
A: Amazon shopping cart (Dynamo), DNS propagation, social media like counts, product catalog updates.

---

**Q: How is eventual consistency achieved technically?**
A: Async replication, no read coordination, anti-entropy (Merkle trees) for background repair, gossip for membership.

---

**Q: When is eventual consistency the wrong model?**
A: When stale reads cause correctness violations — bank balance, inventory count, seat booking.

---

### Session Consistency (Read Your Own Writes)

**Q: What does session consistency guarantee?**
A: You always see your own writes immediately. Other clients may still see stale data.

---

**Q: What is session consistency's CAP choice?**
A: AP (weak) — stays available, but guarantee may break on node failover.

---

**Q: Name three real-world systems that use session consistency.**
A: Twitter (your own tweets), user profile updates, e-commerce order confirmation.

---

**Q: What are the three ways to implement read your own writes?**
A: 1) Sticky sessions (consistent hashing on user ID). 2) Session ID in cookie (load balancer routing). 3) Version tokens (client attaches write version to reads).

---

**Q: What is the failure mode of sticky sessions?**
A: If the assigned node goes down, the user loses the read-your-own-writes guarantee until their session re-establishes.

---

**Q: How do version tokens work?**
A: After a write, server returns a version token. Client sends token with subsequent reads. Any node that has caught up to that version can serve the read — not tied to a specific node.

---

**Q: Which is more resilient to node failures — sticky sessions or version tokens?**
A: Version tokens — any node can serve the read once it has the version. Sticky sessions are tied to one node.

---

### Causal Consistency

**Q: What does causal consistency guarantee?**
A: If write B happened because of write A, everyone sees A before B. Unrelated writes can be reordered freely.

---

**Q: What is causal consistency's CAP choice?**
A: AP (stronger) — stays available, delays causally dependent reads until dependencies are satisfied.

---

**Q: Name three real-world systems that use causal consistency.**
A: Facebook comments (reply after post), Google Docs (edits in response order), WhatsApp message threads.

---

**Q: How is causal consistency implemented?**
A: Vector clocks — client sends its vector clock with every request. Node only serves read after applying all writes the client has seen.

---

**Q: Does Dynamo provide causal consistency?**
A: No. Dynamo uses vector clocks for conflict detection only. A client can still see writes out of causal order. Vector clocks tell you divergence happened — not what order to show writes.

---

**Q: What is the difference between Dynamo's use of vector clocks and causal consistency?**
A: Causal consistency uses vector clocks to delay reads until dependencies are met. Dynamo uses vector clocks to detect conflicts and return all versions to the client — it does not enforce ordering.

---

### Strong Consistency (Linearizability)

**Q: What does strong consistency guarantee?**
A: Every read reflects the most recent write. System behaves as if there is one copy of data and operations happen instantaneously in a global order.

---

**Q: What is strong consistency's CAP choice?**
A: CP — correct during partition, may block or reject operations when quorum unreachable.

---

**Q: Name four real-world systems that require strong consistency.**
A: Bank transfers, flight/concert seat booking, inventory management (can't oversell), distributed locks.

---

**Q: How does Raft achieve strong consistency?**
A: Every write goes through the leader, committed to majority before acknowledging. Reads go through leader. No stale reads possible.

---

**Q: Why can't a Raft follower serve reads directly?**
A: The follower may not have the latest committed entries yet — serving from it could return stale data, violating linearizability.

---

**Q: What is two-phase commit?**
A: Coordinator asks all nodes to prepare → all lock and confirm ready → coordinator commits. Used when a write spans multiple nodes/shards.

---

**Q: What is the risk of two-phase commit?**
A: Coordinator failure leaves all nodes locked — system blocks until coordinator recovers.

---

**Q: What is the difference between Raft quorum and single-leader synchronous replication?**
A: Raft commits when majority acknowledges — one slow node doesn't block all writes. Synchronous replication waits for ALL nodes — stronger durability but one slow node blocks everything.

---

### Choosing the Right Model

**Q: Shopping cart — which consistency model?**
A: Eventual. Stale cart briefly is acceptable. Always-writable is the priority. Data type supports union merge.

---

**Q: Bank account balance — which consistency model?**
A: Strong. Stale reads cause incorrect balance. Cannot tolerate data loss on concurrent writes.

---

**Q: Live comment section where reply must appear after original post — which model?**
A: Causal. Reply ordering within a thread must be preserved. Global ordering across all comments not required.

---

**Q: User logout / session invalidation — which model?**
A: Strong or session. A stale replica showing a logged-out user as still logged in is a security violation.

---

**Q: Game leaderboard updating every 5 minutes — which model?**
A: Eventual. Slightly stale rankings are acceptable. High read throughput is more important.

---

**Q: Concert ticket inventory at sale time — which model?**
A: Strong. Two users cannot both buy the last ticket. Eventual consistency here causes overselling.

---

### CAP & Partitions

**Q: What happens to an AP system during a partition?**
A: Both sides keep serving requests — possibly with stale data. After partition heals, replicas must reconcile.

---

**Q: What happens to a CP system during a partition?**
A: The minority partition stops serving requests. Only the majority partition continues. Availability is sacrificed for correctness.

---

**Q: Why is CA (consistent + available, no partition tolerance) not realistic at scale?**
A: Network partitions always happen at scale. You cannot assume the network is reliable, so partition tolerance is not optional.

---

**Q: After a partition heals in an AP system, what must happen?**
A: Diverged replicas must reconcile — anti-entropy (Merkle trees), conflict resolution (vector clocks or LWW), and gossip to propagate updated membership.
