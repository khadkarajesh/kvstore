Classic Patterns — Flashcards

Format: Q → A
Cover the answer, recall it, then check.

---

**Q: What is base62 encoding and why is it used for URL shortener codes?**
A: Alphabet of 62 characters: [a-z A-Z 0-9]. Encode a numeric ID (or hash) in base62 to get a short, URL-safe string. 7 characters → 62^7 = 3.5 trillion unique codes. Used because it is compact, readable, and avoids special characters that need URL-escaping.

---

**Q: 301 vs 302 redirect in a URL shortener — what is the tradeoff?**
A: 301 (Permanent): browser caches the redirect → subsequent clicks go directly to destination, never hitting your servers again. Fast, cheap, but analytics tracking breaks (requests bypass you). 302 (Temporary): browser asks your server every time → you can log every click and update the destination. Use 301 for pure scale, 302 when analytics are required.

---

**Q: Why is a sequential integer ID a bad URL shortener code? What is the fix?**
A: Sequential IDs are enumerable — an attacker can scrape all shortened URLs by incrementing from 1. Also reveals business volume (code /1000000 means ~1M links created). Fix: use a distributed monotonic ID (Snowflake-style) encoded in base62 → unpredictable ordering, still globally unique.

---

**Q: Fan-out on write vs fan-out on read — state the core tradeoff in one sentence each.**
A: Fan-out on write: on post, push postID to all followers' feed caches → O(1) read, O(followers) write. Fan-out on read: on read, pull from all followees and merge → O(1) write, O(followees) read.

---

**Q: What is the hybrid feed model and what threshold triggers the switch?**
A: Normal users (< 1M followers) → fan-out on write, feed pre-assembled in Redis. Celebrities (≥ 1M followers) → fan-out on read, their posts pulled and merged at read time. Read path combines: pre-assembled cache + live celebrity post fetch + merge sort.

---

**Q: What is the celebrity / hotspot problem in a news feed and why does it matter?**
A: A user with 100M followers who posts triggers 100M Redis writes. At fan-out on write, this saturates write workers for minutes and delays all other users' feeds. Ignoring this in an interview signals you've never thought about skewed distributions at scale.

---

**Q: Why WebSockets for chat instead of HTTP polling?**
A: HTTP polling: client asks every N seconds → up to N seconds delay, wastes bandwidth. WebSocket: persistent bidirectional TCP connection. Server pushes messages instantly with no polling overhead. Tradeoff: stateful connection is pinned to a specific server → harder to load balance, requires sticky sessions or connection registry.

---

**Q: What are per-conversation sequence numbers and what problem do they solve?**
A: A monotonically increasing counter assigned to each message within a conversation. Ensures deterministic ordering regardless of arrival order at the server or client. Client detects gaps (e.g., sequence jumps from 5 to 7) and requests the missing message from the server.

---

**Q: What happens to messages when the chat recipient is offline? Walk through the mechanism.**
A: Message stored in messages table + inserted into recipient's inbox table (undelivered messages). When recipient reconnects, client fetches inbox → delivers all pending messages → deletes inbox rows → sends delivery receipts. For device-offline: push notification via APNs (iOS) or FCM (Android) carries a preview; full message fetched on app open.

---

**Q: Why is synchronous notification sending in the triggering service a mistake?**
A: Notification sends involve external providers (APNs, Twilio, SendGrid) with latency of 0.1s–30s. A provider outage would cascade into the main service being unavailable. Fan-out to millions of followers is O(N) on the hot path. Fix: publish event to a message queue; notification workers consume asynchronously, decoupled from the write path.

---

**Q: What is an idempotency key in a notification system and why is it necessary?**
A: A deterministic unique key per notification, typically hash(user_id + event_id + channel). Stored in Redis after a successful send. Before each send attempt, worker checks Redis — if key exists, skip (already sent). Required because at-least-once delivery guarantees workers will retry on failure; without idempotency keys, retries produce duplicate notifications.

---

**Q: BFS vs DFS for a web crawler — which is preferred and why?**
A: BFS. DFS risks getting trapped in deep link chains (infinite pagination, generated parameter URLs) and monopolizes one domain. BFS spreads across many domains, discovers important pages (many inbound links) sooner, and naturally distributes crawl work. All major crawlers (Googlebot, Bingbot) use BFS.

---

**Q: What is a bloom filter and how does it help a web crawler?**
A: Probabilistic set membership structure. Space: ~10 bits per element → 1B URLs ≈ 1.25GB in memory. Lookup: O(k) hash functions, no disk. False positives possible (filter says "seen" but URL is new → skipped). False negatives impossible ("unseen" is always correct). Enables O(1) near-memory deduplication for 1B+ URLs without a DB round trip per URL.

---

**Q: What must a crawler do before fetching any URL from a domain it hasn't visited?**
A: Fetch and parse robots.txt at the domain root. Cache it (TTL: hours). Check the requested URL path against Disallow rules. Respect the Crawl-delay directive. Never crawl disallowed paths. Violating robots.txt risks IP bans and is considered hostile crawling.

---

**Q: Availability — 99.9% downtime per year, per month, per week.**
A: Per year: ~8.77 hours. Per month: ~43.8 minutes. Per week: ~10.1 minutes.

---

**Q: Availability — 99.99% downtime per year, per month, per week.**
A: Per year: ~52.6 minutes. Per month: ~4.38 minutes. Per week: ~1.01 minutes.

---

**Q: What is the rule of thumb for how availability changes with each additional 9?**
A: Each additional 9 reduces allowed downtime by approximately 10x. 99% → 99.9% → 99.99% → 99.999%: downtime goes 3.65 days → 8.77 hours → 52 minutes → 5.26 minutes.

---

**Q: SLI vs SLO vs SLA — define each in one sentence.**
A: SLI: the measured metric (e.g., "99.94% of requests succeeded last month"). SLO: the internal target engineering commits to (e.g., "we target 99.9% monthly"). SLA: the contractual promise to customers with penalties if breached (e.g., "we guarantee 99.5%, credits if missed").

---

**Q: What is an error budget and what policy applies when it is exhausted?**
A: Error budget = 100% - SLO. At 99.9% SLO: 0.1% budget = 8.77 hours/year. Policy when exhausted: freeze all non-reliability feature releases. Only reliability improvements and critical bug fixes may be deployed until the budget replenishes next month. Prevents the "move fast vs stay reliable" argument by making it a data-driven rule.

---

**Q: Why does p99 latency matter more than p50 at scale?**
A: p50 = median, half of requests. p99 = the slowest 1 in 100 requests. At 1M QPS: p99 = 10,000 users/second experiencing the slow tail. Average masks tail: 990 fast requests + 10 very slow ones can produce a "fine" average while 1% of users see 5-second waits. SLOs and alerts must target p99 (or p999), not average.

---

**Q: Sharding vs replication — state the distinction precisely.**
A: Replication: the same data stored on multiple nodes → fault tolerance and read scale. Sharding: different data on each node → write scale and storage scale. They are orthogonal. Production systems use both: each shard has multiple replicas. Confusing them in an interview is a red flag.

---

**Q: What is the core weakness of hash(key) % N sharding when you need to add a shard?**
A: Adding one shard changes N to N+1. Almost all keys map to different shards: hash(key) % N ≠ hash(key) % (N+1) for ~(N/(N+1)) of keys → ~80-90% of data must move to new shards. Solution: consistent hashing — only K/N keys move on node add/remove (1/N of total keys).

---

**Q: What is the core weakness of range-based sharding and what causes it?**
A: Hotspot problem. If the shard key is sequential (auto-increment IDs, timestamps), all new writes go to the last shard — the range containing the highest values. All other shards are idle. Fix: add a hash prefix to distribute writes, or use consistent hashing instead of range boundaries.

---

**Q: Name three solutions to the hot shard / hot key problem.**
A: 1) Key prefix salting: write key to N shards with salt prefix (key:0, key:1, ..., key:N), read from all and merge. 2) Split the hot shard: detect via monitoring, split its range into two, migrate half the data. 3) Dedicated shard for celebrity entities: identify top-K hot keys, assign them dedicated shard(s) with application-level awareness.

---

**Q: What is the cost of a cross-shard JOIN and how do you avoid it?**
A: Cross-shard JOIN requires scatter-gather: query sent to all shards, each returns partial results, application merges. Cost: O(shards) latency × O(result size) data transfer. Avoid by designing the shard key to co-locate data queried together — e.g., shard social data by user_id so all of user X's posts, follows, and likes are on the same shard.

---

**Q: Senior vs junior signal — what is the single clearest behavioral difference in a system design interview?**
A: Junior: describes what a system does and picks technologies. Senior: explains why each choice was made, what was ruled out, what the tradeoff costs — and raises failure modes and tradeoffs unprompted before the interviewer asks. The senior candidate leads the interview; the junior candidate answers questions.

---
