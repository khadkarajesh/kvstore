## Sharding — Flashcards

Format: Q → A
Cover the answer, recall it, then check.

---

**Q: What is sharding and what problem does it solve?**
A: Sharding is horizontal partitioning — splitting a dataset across multiple machines so each machine owns a distinct subset of the data. It solves write throughput limits and storage limits that a single machine cannot handle, unlike replication which only helps with read scale and fault tolerance.

---

**Q: What is the critical distinction between sharding and replication?**
A: Replication: same data on multiple nodes — provides fault tolerance and read scale. Sharding: different data on each node — provides write scale and storage scale. Production systems use both: each shard has multiple replicas. Confusing these two in an interview is a red flag.

---

**Q: When should you shard? What should you exhaust first?**
A: Exhaust vertical scaling and read replicas before introducing sharding. Shard when: single machine cannot hold all data in storage, write throughput exceeds what a single node can handle, or read throughput exceeds what replicas can handle.

---

**Q: How does range-based sharding work?**
A: Keys are sorted and the keyspace is split into contiguous ranges. Each range is assigned to exactly one shard. Routing requires finding which range contains the key via binary search over sorted range boundaries — O(log S) where S is the number of shards.

---

**Q: What is the main advantage of range-based sharding?**
A: Range scans are efficient — all rows in a range land on the same shard, so SELECT WHERE key BETWEEN x AND y is served by one shard with no scatter-gather needed.

---

**Q: What is the main disadvantage of range-based sharding?**
A: Hotspot problem. If keys are sequential (timestamps, auto-increment IDs), all new writes go to the last shard while all other shards are idle. Also, some ranges may contain far more data than others, causing shard imbalance.

---

**Q: How does hash-based sharding work?**
A: shard = hash(key) % num_shards. Keys are uniformly distributed across shards regardless of their ordering. Routing is O(1) — compute the hash, apply modulo, done. No lookup table needed.

---

**Q: What is the catastrophic resharding problem with hash(key) % N?**
A: Adding one shard changes N to N+1, so hash(key) % (N+1) is recomputed for all keys. Roughly (N-1)/N of all keys must move to different shards. For 4 → 5 shards: ~80% of all keys must migrate. This is data migration at massive scale.

---

**Q: Why are range scans broken with hash-based sharding?**
A: Adjacent keys hash to different shards, so a range query must be scattered to all shards and results gathered in the application layer. There is no way to serve a range query from a single shard.

---

**Q: How does directory-based sharding work?**
A: A lookup service (directory) maps each key to a shard ID. The client or proxy queries the directory before each request to find the target shard.

---

**Q: What are the main risks of directory-based sharding?**
A: The directory is a single point of failure — if it goes down, the entire system cannot route. It is also a bottleneck — every request requires a directory lookup (mitigated by client-side caching with TTL). Stale cache entries after a key moves require careful invalidation.

---

**Q: How does consistent hashing solve the resharding problem?**
A: The hash space forms a ring [0, 2^32). Both keys and nodes are hashed onto this ring. A key is assigned to the first node clockwise from its position. Adding one node moves only K/N keys (1/N of total). Simple modulo hashing moves ~(N-1)/N of keys — catastrophic by comparison.

---

**Q: What are virtual nodes in consistent hashing and why are they used?**
A: Each physical node is represented by 50-150 virtual nodes scattered across the ring. Benefits: (1) even distribution despite random physical node placement, (2) a failed node's keys spread across many physical successors rather than overloading one, (3) heterogeneous hardware — stronger machines get more virtual nodes for proportional load.

---

**Q: What is a hot shard and what causes it?**
A: A shard that receives disproportionate traffic or stores disproportionately large data while other shards are underutilized. Causes: skewed key access (most queries hit one key), celebrity keys (Taylor Swift's tweet accessed 100M times more than average), temporal keys (range-sharded time-series where all writes go to "today" shard), auto-increment IDs with range sharding.

---

**Q: How does key prefix salting eliminate a celebrity key hotspot?**
A: Add a random salt prefix in [0, N): write to tweet:0:tswift, tweet:1:tswift, ..., tweet:N-1:tswift across all N shards. Reads scatter to all N shards and merge (sum for counters, max for latest). Tradeoff: write amplification x N and read scatter x N — acceptable only for a small number of very hot keys.

---

**Q: What are the steps for online range shard migration?**
A: (1) New shard starts receiving writes — dual-write to old and new shard. (2) Backfill: copy historical data from old shard to new shard. (3) Verify: confirm new shard has all data. (4) Cutover: redirect reads and writes to new shard only. (5) Drain and decommission old shard.

---

**Q: How do cross-shard JOINs work and how do you avoid them?**
A: Cross-shard JOINs require scatter-gather: send query to all shards, each returns a partial result, then merge in the application layer. Cost is O(shards) in latency and O(shards × result_size) in data transfer. Avoid by denormalization — duplicate the joined data into one table so the JOIN is never needed.

---

**Q: What is Two-Phase Commit (2PC) and why is it avoided in sharded systems?**
A: 2PC is a cross-shard atomicity protocol: coordinator sends PREPARE to all participant shards; if all reply READY, coordinator sends COMMIT; if any ABORT, coordinator sends ROLLBACK. Problems: coordinator failure during phase 2 leaves participants blocked; network partition can cause some shards to commit and others to abort; 2 round trips to all shards before completion. Production approach: design the shard key to co-locate related data so 2PC is never needed.

---

**Q: What is the shard key design principle for avoiding cross-shard operations?**
A: Shard key should co-locate data that is queried or written together. Example: in a social network, shard by user_id so all of a user's posts, likes, and follows land on the same shard. Sharding posts by post_id and users by user_id forces a cross-shard JOIN for every "user's posts" query.

---

**Q: In a sharded system with replication, what determines total node count and what failures can be tolerated?**
A: Total nodes = num_shards x replication_factor. Example: 10 shards x 3 replicas = 30 nodes. Up to (replication_factor - 1) node failures per shard can be tolerated. Writes go to the primary of the correct shard and are replicated to its replicas. Reads can be served by any replica of the correct shard.

---
