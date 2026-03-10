Caching — System Design Notes

---
Why Caching Exists

Latency:   disk read ~1ms, network DB read ~10ms, cache read ~0.1ms (Redis), local cache ~100ns
Throughput: DB can serve ~10K QPS, Redis ~100K QPS, local cache unlimited (no network)
Cost:       DB instances are expensive; cache amortizes load across fewer DB nodes
Observation: 80% of reads often hit 20% of data → caching that 20% eliminates 80% of DB load

---
Cache Placement Strategies

Cache-Aside (Lazy Loading)
  Flow:
    Read: app checks cache → hit → return | miss → read DB → write to cache → return
    Write: app writes DB → invalidate or update cache entry
  When to use: read-heavy, irregular access patterns, tolerable stale window
  Tradeoffs:
    + cache only holds data that is actually requested
    + DB failure degrades gracefully (cache still serves)
    - stale data window exists between write and cache update
    - cold start: first request always hits DB (cache miss storm on restart)
    - app owns cache logic → more application code complexity

Write-Through
  Flow:
    Write: app writes cache → cache proxy writes DB synchronously → both updated before ack
    Read: check cache → hit → return | miss → read DB → populate cache → return
  When to use: write-heavy but read-dominated after write, can afford write latency, zero stale tolerance
  Tradeoffs:
    + cache always consistent with DB — no stale window
    + reads after write always hit cache
    - write latency doubles (two hops: cache + DB)
    - pollutes cache with data that may never be read again

Write-Behind (Write-Back)
  Flow:
    Write: app writes cache → cache acks immediately → cache async flushes to DB later
    Read: same as write-through
  When to use: write-intensive workloads where DB throughput is the bottleneck
  Tradeoffs:
    + lowest write latency (no DB on write path)
    + batches DB writes → fewer IOPS
    - data loss window if cache node crashes before flush
    - DB can lag cache → other readers hitting DB see stale data
    - complex: need durable queue or WAL to survive cache failure

Write-Around
  Flow:
    Write: app writes directly to DB, bypasses cache entirely
    Read: cache miss → read DB → optionally populate cache
  When to use: write-once read-rarely data (audit logs, archives), large infrequently read objects
  Tradeoffs:
    + prevents cache pollution from write-heavy data
    - first read always misses cache (cold read)
    - high cache miss rate if reads follow writes quickly

Strategy Selection Guide
  Tolerate stale, read-heavy              → cache-aside
  Zero stale, write then read pattern     → write-through
  Write-heavy, can tolerate data loss     → write-behind
  Write-once, read-rarely                 → write-around

---
Cache Eviction Policies

LRU — Least Recently Used
  Mechanism: evict the key that was accessed longest ago
  Implementation: doubly linked list + hash map → O(1) get + O(1) put
  When to use: temporal locality — recent data likely to be requested again
  Weakness: one large scan pollutes the entire cache (cache pollution)

LFU — Least Frequently Used
  Mechanism: evict the key with lowest access count; break ties by recency
  Implementation: frequency → doubly linked list of keys at that frequency; O(1) with min pointer
  When to use: skewed access — some keys are permanently hot (user profiles, config, top-N items)
  Weakness:
    - historical frequency bias: old hot items survive even if no longer needed
    - new items start at count=1 → evicted before proving value (cold start bias)

TTL — Time To Live
  Mechanism: each key expires after a fixed duration regardless of access pattern
  When to use:
    - data has natural expiry (sessions, rate limits, auth tokens)
    - forcing freshness is more important than hit rate
  Weakness:
    - does not consider access frequency or recency
    - cache avalanche risk when many keys share same TTL

Combined:  most production caches use LRU or LFU + TTL together
           LRU for eviction order, TTL as a correctness bound on staleness

---
Redis Architecture

Single-Threaded Event Loop
  Redis is single-threaded for command processing
  → no lock contention between commands — each command is atomic
  → no context switching overhead
  → throughput: ~100K ops/sec on commodity hardware
  Exception: background I/O threads for persistence (AOF flush, RDB snapshot, network writes)
  Redis 6.0+ added multi-threaded I/O for reading/writing network sockets
  → parsing requests is multi-threaded; command execution remains single-threaded

Key Data Structures

  String
    - Binary safe, max 512MB
    - Use case: session tokens, counters (INCR is atomic), feature flags
    - Commands: GET, SET, INCR, DECR, SETNX (set if not exists)

  Hash
    - Map of field → value, all under one key
    - Use case: user profile (user:123 → {name: "alice", email: "...", age: 30})
    - More memory-efficient than one key per field
    - Commands: HGET, HSET, HMGET, HGETALL

  List
    - Doubly linked list, O(1) push/pop from both ends
    - Use case: message queue, activity feed, recent N items (LPUSH + LTRIM)
    - Commands: LPUSH, RPUSH, LPOP, RPOP, LRANGE, LTRIM

  Set
    - Unordered collection of unique strings
    - Use case: unique visitors per page, user followers (SINTERSTORE for mutual follows)
    - O(1) add, remove, check membership
    - Commands: SADD, SISMEMBER, SINTER, SUNION, SDIFF

  Sorted Set (ZSet)
    - Each member has a score; members sorted by score
    - Use case: leaderboards (ZADD user:scores 100 "alice"), rate limiting, priority queues
    - Range queries by score or rank: ZRANGE, ZRANGEBYSCORE, ZRANK
    - O(log N) operations (backed by skip list + hash map)

Persistence

  RDB (Redis Database Snapshot)
    - Point-in-time snapshot of entire dataset written to disk
    - Triggered: BGSAVE or schedule (save 900 1 = save if 1 key changed in 900s)
    - Fork-on-write: parent keeps serving; child writes snapshot to disk
    - Tradeoffs:
        + compact binary file, fast restart
        + low overhead during normal operation
        - data loss window = time since last snapshot (minutes)
        - fork can cause latency spike on large datasets

  AOF (Append-Only File)
    - Every write command appended to log file before ack
    - fsync options:
        always:  fsync after every command → zero data loss, ~1K QPS
        everysec: fsync every second → max 1s data loss (default)
        no:       OS decides when to flush → fastest, no durability guarantee
    - AOF Rewrite: compacted periodically (BGREWRITEAOF) — replays produce same state
    - Tradeoffs:
        + far less data loss than RDB
        + AOF file is human-readable, recoverable
        - larger file size, slower restart (replay vs load)

  Recommendation: RDB + AOF together → fast restart from RDB, replay AOF delta for recent writes

Redis Cluster — Sharding

  Hash slots: key space divided into 16384 slots
    slot = CRC16(key) mod 16384
  Each master node owns a range of slots (e.g. node1: 0-5460, node2: 5461-10922, node3: 10923-16383)
  Client hashes key → knows which node to talk to directly (no proxy)
  Adding a node: slots migrated from existing nodes to new node; zero downtime
  Replication: each master has N replicas; replica promoted on master failure

  Hash tags: {user}.profile and {user}.cart → same slot (bracket extracts key for hashing)
    → enables multi-key operations across logically related keys

Redis Sentinel — High Availability (non-cluster)

  Sentinel is a separate process monitoring Redis master + replicas
  Responsibilities:
    - Monitoring: heartbeat checks to master
    - Leader election: Sentinels elect a leader via Raft-like quorum (majority)
    - Failover: leader Sentinel promotes a replica to master, updates clients
    - Config propagation: notifies clients of new master address
  Quorum: need majority of Sentinels to agree master is down (avoids split-brain)
  Typical deploy: 3 Sentinels, 1 master, 2 replicas

---
Cache Problems and Solutions

Cache Invalidation
  The hardest problem in caching: how to keep cache consistent with DB
  Strategies:
    TTL-based:        cache expires after fixed time → stale window = TTL duration
    Write-through:    update cache on every write → strong consistency, write latency
    Event-driven:     DB change → pub/sub event → invalidate cache key
    Version-based:    include version in cache key (user:123:v5) → old version naturally evicted

Thundering Herd (Cache Stampede)
  What: popular key expires → N concurrent requests all miss cache simultaneously
        → N requests all hit DB at once → DB overwhelmed
  Solutions:
    Mutex lock:                  first miss acquires lock, fetches DB, populates cache
                                 others wait for lock → serialize DB hits to 1
    Probabilistic early expiry:  before key expires, each request has small probability of
                                 triggering refresh early → spreads DB load over time
    Request coalescing:          collapse N in-flight requests for same key into one DB call
                                 N callers all get the same response when it returns

Hot Key Problem
  What: one key (e.g. viral tweet, trending item) receives millions of requests/sec
        → single Redis shard overwhelmed
  Solutions:
    Local L1 cache:    serve hot key from in-process cache → no network hop
                       stale window = L1 TTL (seconds), acceptable for viral content
    Key replication:   copy hot key to multiple shards (tweet:123:shard0 ... shard9)
                       client reads from random shard → distribute load
    Key splitting:     hot counter → split into N shards, sum on read

Cache Penetration
  What: requests for keys that don't exist in cache OR DB → every request hits DB
        common attack vector: query for non-existent user IDs
  Solutions:
    Cache null values:  store "NULL" sentinel in cache with short TTL
                        → subsequent requests hit cache, not DB
    Bloom filter:       probabilistic structure, checked before cache
                        "definitely not in DB" → return immediately, no cache or DB hit
                        false positives possible but false negatives impossible

Cache Avalanche
  What: large number of keys expire at the same time → massive DB hit spike
  Why it happens: batch load with same TTL, Redis restart clears all keys
  Solutions:
    TTL jitter:          TTL = base_ttl + random(0, jitter)
                         spread expiry over a window instead of single moment
    Staggered reload:    background job pre-warms cache before expiry
    Circuit breaker:     limit DB request rate during avalanche, serve stale or error

---
Local Cache vs Distributed Cache

Local Cache (in-process)
  - Latency: microseconds (CPU cache + RAM, no syscall)
  - No network: zero serialization overhead
  - Size: bounded by process memory (typically 100MB-1GB)
  - Consistency: each node has its own copy → inconsistent across replicas
  - Failure: lost on process restart
  - Examples: Guava LoadingCache, Caffeine (Java), sync.Map (Go), functools.lru_cache (Python)
  When to use: hot config, reference data, computation cache where stale is acceptable

Distributed Cache (Redis, Memcached)
  - Latency: 0.1ms–1ms (network hop + serialization)
  - Network: one round trip per read/write
  - Size: scales horizontally across nodes (TBs possible)
  - Consistency: all app nodes share same view (single source of truth)
  - Failure: survived independently of app nodes; replicated for HA
  - Examples: Redis, Memcached, Hazelcast
  When to use: shared session state, rate limiting, distributed locks, real-time leaderboards

Two-Tier Caching (L1 + L2)
  Architecture: app → L1 local cache → L2 Redis → DB
  Flow:
    L1 hit → return (microseconds)
    L1 miss → L2 hit → populate L1 → return (~1ms)
    L2 miss → DB read → populate L2 + L1 → return (~10ms)
  L1 TTL: short (1-5s) to bound staleness across replicas
  Use cases: Twitter timeline rendering, product catalog, ad targeting
  Tradeoff: adds staleness window = L1 TTL, worth it for eliminating Redis on every request
