## Caching — Flashcards

Format: Q → A
Cover the answer, recall it, then check.

---

**Q: Walk through the cache-aside read and write flow.**
A: Read: check cache → hit → return | miss → read DB → write to cache → return. Write: write to DB → invalidate (or update) cache key. App owns all cache logic.

---

**Q: What is the stale data weakness of cache-aside?**
A: After a write to DB, the cache still holds the old value until invalidation or TTL expiry. Any read in that window returns stale data. The stale window = time between write and cache update.

---

**Q: Write-through vs write-behind — what is the core tradeoff?**
A: Write-through: writes DB and cache synchronously → zero stale window, double write latency. Write-behind: writes cache only, DB async → lowest write latency, data loss risk if cache crashes before flush.

---

**Q: When would you choose write-around over write-through?**
A: When data is written once and read rarely (audit logs, backups, archives). Write-through would pollute the cache with data that will never generate cache hits.

---

**Q: LRU vs LFU — which one handles a one-time sequential scan better?**
A: LFU. LRU evicts everything currently in cache (scan pollution — scan items become the most recently used). LFU keeps genuinely hot items because they have high frequency counts; scan items start at count=1.

---

**Q: LRU vs LFU — which one handles permanent hot items better?**
A: LFU. Items accessed millions of times accumulate high frequency counts and survive eviction. LRU would evict them if they haven't been accessed recently (e.g. after a traffic lull).

---

**Q: What is the cold start problem with LFU?**
A: New items start at frequency count = 1 — the same as the minimum. They are evicted immediately before accumulating hits. A genuinely popular new item may be evicted before it proves its value.

---

**Q: Redis String — give one use case.**
A: Session tokens (SET session:abc123 "user:42" EX 3600), or atomic counters (INCR page:views returns the new value atomically, safe without transactions).

---

**Q: Redis Hash — give one use case.**
A: User profile: HSET user:123 name "alice" email "a@b.com" age 30. All fields under one key — more memory-efficient than one string key per field, and HGETALL returns the full object in one round trip.

---

**Q: Redis List — give one use case.**
A: Activity feed or simple queue: LPUSH feed:user:123 "event" + LTRIM feed:user:123 0 99 keeps the most recent 100 events. RPOP for queue consumption (FIFO).

---

**Q: Redis Set — give one use case.**
A: Unique visitors: SADD page:views:2024-03-10 user:42 — Set automatically deduplicates. SINTERSTORE mutual_follows user:42:following user:99:following — finds common follows in one command.

---

**Q: Redis Sorted Set — give one use case.**
A: Leaderboard: ZADD game:scores 9800 "alice" 8500 "bob". ZRANGE game:scores 0 9 REV WITHSCORES returns top-10. Backed by skip list — O(log N) insert, O(log N + K) range query.

---

**Q: What is the thundering herd problem and what triggers it?**
A: A popular cache key expires → many concurrent requests all miss cache simultaneously → all hit the DB at once → DB overwhelmed. Triggered by: key expiry, cache restart, cold start.

---

**Q: Name three solutions to thundering herd.**
A: 1) Mutex lock — first miss locks, fetches DB, populates cache; others wait. 2) Probabilistic early expiry — random chance of refreshing before expiry, spreads load. 3) Request coalescing — collapse N in-flight requests for same key into one DB call.

---

**Q: What is the hot key problem and what is its root cause?**
A: One key (viral tweet, trending product) receives millions of req/sec → single Redis shard becomes the bottleneck. Root cause: keyspace is non-uniform, consistent hashing puts one logical key on one shard.

---

**Q: Name two solutions to the hot key problem.**
A: 1) Local L1 cache — serve hot key from in-process cache, no Redis hop, stale by L1 TTL. 2) Key replication — copy key to N shards (tweet:123:shard0…shardN), read from random shard to distribute load.

---

**Q: Cache penetration vs cache avalanche — what is the difference?**
A: Penetration: requests for keys that don't exist anywhere → every request bypasses cache and hits DB (often a DoS attack). Avalanche: many valid keys expire simultaneously → mass DB hit spike.

---

**Q: What are two solutions to cache penetration?**
A: 1) Cache null values — store a "NULL" sentinel with a short TTL so subsequent requests for the same non-existent key hit cache. 2) Bloom filter — probabilistic check before cache; "definitely not in DB" returns immediately without touching cache or DB.

---

**Q: What is the solution to cache avalanche?**
A: TTL jitter — set TTL = base_ttl + random(0, jitter) so keys expire spread over a window instead of all at once. Also: background pre-warming before expiry, circuit breaker to rate-limit DB during an avalanche.

---

**Q: Redis hash slots — how many, and how is a key assigned to one?**
A: 16384 hash slots total. Assignment: slot = CRC16(key) mod 16384. Each master node owns a contiguous range of slots. Clients compute the slot locally and talk directly to the right node — no proxy required.

---

**Q: What are Redis hash tags and why do they exist?**
A: Curly braces in a key name: {user:42}.profile and {user:42}.cart both hash only the {user:42} portion → guaranteed same slot → multi-key commands (MGET, transactions) work without cross-slot errors.

---

**Q: Redis RDB vs AOF — which has lower data loss risk?**
A: AOF (with fsync=everysec) — max 1 second of data loss. RDB snapshots can be minutes apart; a crash loses all writes since the last snapshot.

---

**Q: What is the recommended Redis persistence configuration in production?**
A: RDB + AOF together. Fast restart from the compact RDB snapshot, then replay AOF for recent writes that happened after the snapshot. Neither alone is optimal.

---

**Q: Local cache vs distributed cache — when do you choose local?**
A: When latency must be microseconds (not milliseconds), the data is small, per-node inconsistency is acceptable (e.g. config, reference data, computation results), and there is no need to share state across replicas.

---

**Q: What is two-tier caching and what problem does it solve?**
A: L1 local cache (in-process) + L2 distributed cache (Redis) + DB. Eliminates Redis network hop for the hottest requests. L1 TTL (1-5s) bounds cross-replica staleness. Used when Redis itself becomes the bottleneck at very high QPS.

---
