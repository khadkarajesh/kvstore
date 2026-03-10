Sharding — System Design Notes

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
What Sharding Is and Why It Exists
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Sharding = horizontal partitioning: split a dataset across multiple machines
             so each machine owns a distinct subset of the data.

  Why not vertical scaling?
    Vertical: buy a bigger machine (more RAM, faster CPU, faster disk)
    Limits: hardware ceiling exists. 192-core, 24TB RAM machines exist but cost $500K+.
    Operational risk: one machine → one point of failure regardless of specs.
    Network I/O: single machine's NIC becomes bottleneck before CPU does at high QPS.

  Sharding vs Replication — critical distinction:
    Replication: same data on multiple nodes → fault tolerance + read scale
    Sharding:    different data on each node → write scale + storage scale
    Production systems use both: each shard has multiple replicas.
    Confusing these two in an interview is a red flag.

  When to shard:
    Single machine cannot hold all data in storage → must shard.
    Write throughput exceeds what a single node can handle → must shard.
    Read throughput exceeds what replicas can handle (N replicas still not enough) → must shard.
    Rule: exhaust vertical scaling and read replicas before introducing sharding.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Range-Based Sharding
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Mechanism:
    Keys are sorted. Keyspace split into contiguous ranges.
    Each range assigned to exactly one shard.
    Example:
      Shard 1: user_id 1 – 1,000,000
      Shard 2: user_id 1,000,001 – 2,000,000
      Shard 3: user_id 2,000,001 – 3,000,000

  Routing: route(key) → find which range contains key → target shard.
    Lookup: binary search over sorted range boundaries. O(log S) where S = number of shards.
    Range boundaries stored in a small metadata table or in-memory.

  Advantages:
    Range scans are efficient: all rows in a range land on the same shard.
    No scatter-gather for range queries (SELECT * WHERE user_id BETWEEN 1M AND 2M).
    Predictable data placement: easy to reason about which shard holds what.

  Disadvantages:
    Hotspot problem: if keys are sequential (timestamps, auto-increment IDs), all new writes
      go to the last shard → single shard is the write bottleneck while all others are idle.
    Uneven sizes: some ranges may contain far more data than others → shard imbalance.
    Manual rebalancing: splitting an overloaded shard requires range boundary adjustment.

  Used by: Bigtable (tablets), HBase, TiKV, DynamoDB (sort key within a partition)

  When to choose: workloads with frequent range scans (analytics, time-series, sorted leaderboards).
  Pair with: artificial key prefix (e.g., hash(user_id) % 100 as prefix) to break sequential hotspot.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Hash-Based Sharding
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Mechanism:
    shard = hash(key) % num_shards
    Keys are uniformly distributed across shards regardless of key ordering.
    Example (4 shards):
      hash("user:42")  % 4 = 2  → Shard 2
      hash("user:43")  % 4 = 3  → Shard 3
      hash("user:1M")  % 4 = 1  → Shard 1

  Routing: O(1) — compute hash, apply modulo, done. No lookup table needed.

  Advantages:
    Even distribution: good hash function → uniform keyspace distribution → no natural hotspots.
    Simple implementation: one formula, no metadata.

  Disadvantages:
    Range scans are broken: adjacent keys hash to different shards → range query must scatter to all shards.
    Resharding is catastrophic: adding shard N+1 → all hash(key) % (N+1) recomputed → almost all keys move.
      Example: 4 → 5 shards: 4/5 = 80% of all keys must move. This is data migration at scale.
    Hash collision is irrelevant but key routing is deterministic → easier to DoS by crafting keys.

  Used by: DynamoDB (partition key hashing), Redis Cluster (CRC16 mod 16384 = hash slots), Cassandra

  Resharding solution: consistent hashing (see below) avoids the modulo trap.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Directory-Based Sharding
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Mechanism:
    A lookup service (directory) maps each key → shard ID.
    Client or proxy queries the directory to find the target shard before each request.

  Example:
    Directory: {user:1 → Shard3, user:2 → Shard1, user:3 → Shard2, ...}

  Advantages:
    Maximum flexibility: individual keys can be moved between shards at will.
    Supports heterogeneous shards: move hot keys to more powerful nodes.
    No data movement on add: new shard added → redirect new keys to it, no existing keys move.

  Disadvantages:
    Directory = single point of failure: directory goes down → entire system cannot route requests.
    Directory = bottleneck: every request requires a directory lookup (mitigated by caching).
    Caching the directory → stale entries after a key moves → must invalidate carefully.
    Operationally complex: directory must be maintained, replicated, kept consistent.

  Used by: older sharding frameworks; rarely in modern systems without caching layer on top.
  Mitigation: cache directory mappings in client with TTL, retry on routing failure.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Consistent Hashing
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Solves: the catastrophic resharding problem of hash(key) % N.

  Mechanism:
    Hash space is a ring [0, 2^32). Both keys and nodes are hashed onto this ring.
    Key → assigned to the first node clockwise from the key's hash position.
    Node added: only keys between (new_node, new_node's predecessor] on the ring move.
    Node removed: only that node's keys move to its clockwise successor.

  Key movement on change:
    Adding 1 node to N-node cluster → only K/N keys move (1/N of total keys).
    Simple hash(key) % N: ~(N-1)/N of keys move → catastrophic.

  Virtual nodes:
    Each physical node represented by multiple (50-150) virtual nodes on the ring.
    Benefit: even distribution despite randomness of physical node placement.
    Benefit: node failure spreads its keys across many physical successors (not all onto one).
    Benefit: heterogeneous hardware → strong machines get more virtual nodes → proportional load.

  Full detail → see Dynamo notes (consistent hashing used as Dynamo's partitioning strategy).

  Used by: Amazon Dynamo, Cassandra, Chord, Memcached (ketama).

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Hotspot Problems and Solutions
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  What is a hot shard:
    One shard receives disproportionate traffic or stores disproportionately large data.
    All other shards are underutilized. System bottlenecks on one machine.

  Causes:
    Skewed key distribution: most users query user_id=1 (admin, test account, demo user)
    Celebrity keys: Taylor Swift's tweet ID is accessed 100M× more than average
    Temporal keys: range-sharded time-series → all writes go to the "today" shard
    Auto-increment IDs: sequential → range shard → new rows always on last shard

  Solution 1 — Virtual nodes (consistent hashing)
    Scatter each physical node across the ring → no single point concentrates load.
    Works at the node level; does not help if the hot key is logically one key.

  Solution 2 — Key prefix salting
    Add random salt prefix: key = salt + ":" + original_key, salt ∈ [0, N)
    Write to: tweet:0:tswift, tweet:1:tswift, ..., tweet:N-1:tswift (all N shards)
    Read: scatter read to all N shards → merge result (for counters: sum; for latest: max)
    Tradeoff: write amplification × N; read scatter × N. Acceptable for very hot keys only.

  Solution 3 — Split the hot shard
    Detect overloaded shard via monitoring. Split its range into two shards.
    Works for range-based sharding; temporarily double-write during migration.
    Requires online migration support.

  Solution 4 — Dedicated shard for celebrity entities
    Identify top-K hot keys/entities. Assign dedicated shard (or multiple) just for them.
    Requires application-level awareness of "celebrity" status.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Resharding
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Hash-based resharding:
    Problem: changing N recomputes hash(key) % N for all keys → ~80% of keys must move.
    Solution: consistent hashing → only K/N keys move on node add/remove.
    If using plain modulo sharding: pre-shard with large fixed N (e.g. 1024 virtual shards),
      map virtual → physical shards separately. Adding physical nodes → remaps virtual shards,
      not individual keys. Data still moves but in predictable, large virtual shard units.

  Range-based resharding:
    Split: divide a range into two ranges → assign second half to new shard.
    Online migration:
      1. New shard starts receiving writes (dual-write to old + new shard)
      2. Backfill: copy historical data from old shard to new shard
      3. Verify: confirm new shard has all data
      4. Cutover: redirect reads + writes to new shard only
      5. Drain old shard and decommission
    Requires: dual-write capability in application layer.

  Resharding is operationally expensive. Avoid by:
    Over-provisioning initial shard count.
    Using consistent hashing.
    Starting with fixed virtual shard mapping.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Cross-Shard Operations
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Cross-shard JOINs:
    SQL JOIN across two tables on different shards → requires scatter-gather.
    Scatter: send query to all shards → each returns its partial result.
    Gather: merge results in application layer or query aggregator.
    Cost: O(shards) latency + O(shards × result_size) data transfer.
    Avoid by: denormalization — duplicate the joined data into one table so the join is never needed.

  Cross-shard transactions (2PC — Two-Phase Commit):
    Coordinator sends PREPARE to all participant shards.
    All participants respond READY or ABORT.
    If all READY: coordinator sends COMMIT.
    If any ABORT: coordinator sends ROLLBACK to all.
    Problems:
      - Coordinator failure during phase 2 → participants blocked waiting indefinitely
      - Network partition → some shards commit, others abort → inconsistency
      - Latency: 2 round trips to all shards before the transaction completes
    Production approach: avoid 2PC by designing to keep related data on the same shard.

  Shard key design principle:
    Shard key should co-locate data that is queried or written together.
    Example: social network → shard by user_id. All of user X's posts, likes, follows on same shard.
    → queries about user X never cross shards.
    Counterexample: shard posts by post_id, users by user_id → JOIN for "user's posts" always cross-shard.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Replication + Sharding Together
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Every production sharded system also replicates each shard.

  Architecture:
    Shard 1: [Primary] + [Replica1] + [Replica2]
    Shard 2: [Primary] + [Replica1] + [Replica2]
    Shard 3: [Primary] + [Replica1] + [Replica2]

  Two orthogonal dimensions:
    shard_id → determines which subset of data (which rows/keys)
    replica  → determines which copy of that subset (for fault tolerance + read scale)

  Reads: can be served by any replica of the correct shard.
  Writes: go to the primary of the correct shard → replicated to its replicas.
  Failure: primary of a shard fails → replica promoted → still serving that shard's data.

  Total nodes = num_shards × replication_factor.
  Example: 10 shards × 3 replicas = 30 nodes. Losing up to (replicas-1) nodes per shard = tolerated.

