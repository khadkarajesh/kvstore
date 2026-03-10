Classic System Design Patterns — FAANG Interview Reference

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. URL Shortener (TinyURL)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Core insight: generating a globally unique short code and resolving it at high read throughput
              without enumerable IDs or collision storms.

---
Clarify First

  - Read/write ratio? (100:1 typical — reads dominate → optimize redirect latency)
  - Custom aliases required? (user:123/mylink vs system-generated)
  - Analytics? (click counts, geo, device — or pure redirect?)
  - URL expiry? (TTL per link or permanent)
  - Max URL length supported?

---
Scale Estimation

  100M new URLs/day → ~1,200 writes/sec
  10B redirects/day → ~115,000 reads/sec
  URL row: ~500 bytes avg → 100M × 365 × 500B = ~18TB/year
  → read-heavy, needs heavy caching on hot short codes

---
Key Design Decisions

  Code Generation — base62 encoding
    Alphabet: [a-z A-Z 0-9] = 62 chars
    7 chars → 62^7 = 3.5 trillion unique codes (enough for centuries)

    Option A: Hash-based (MD5/SHA256 → take first 7 chars of base62)
      + deterministic: same long URL → same short code
      - collision possible: different URLs produce same prefix
      - collision handling: append counter suffix, retry until unique

    Option B: Counter-based (global auto-increment ID → encode in base62)
      + guaranteed unique
      - predictable and enumerable (attacker can walk all URLs)
      - global counter = bottleneck at high write rate

    Option C: Random (crypto/rand 7 chars, check DB for collision)
      + unpredictable
      - collision probability rises as DB fills (birthday paradox)
      - check-then-insert requires DB round trip per write

    Production choice: Counter-based with distributed ID generator (Snowflake-style)
      → shard ID embedded in counter → no global bottleneck, still monotonic

  301 vs 302 Redirect
    301 Permanent Redirect: browser caches the mapping → subsequent clicks never hit your server
      + reduces load dramatically
      - cannot track clicks (browser goes directly to destination)
      - cannot update destination without clearing browser cache

    302 Temporary Redirect: browser asks your server every time
      + analytics tracking works (every request is logged)
      + can change destination at any time
      - every redirect hits your servers → need to cache aggressively

    Interview answer: use 302 if analytics required, 301 if pure scale/cost optimization.

  Storage Schema
    Table: urls
    ┌──────────────┬────────────────────┬────────────┬──────────────┐
    │ shortcode    │ long_url           │ created_at │ expires_at   │
    │ VARCHAR(7) PK│ VARCHAR(2048)      │ TIMESTAMP  │ TIMESTAMP    │
    └──────────────┴────────────────────┴────────────┴──────────────┘
    Index on shortcode (primary key) → O(1) lookup
    Optional index on long_url if deduplication (same long → same short) is required

  Redirect Latency — Caching
    Hot codes: 80% of traffic hits 20% of URLs → cache in Redis
    Key: shortcode, Value: long_url, TTL: 24 hours (refresh on hit)
    Cache-aside: miss → DB lookup → populate cache → redirect
    Cache hit rate target: >99% at steady state

  Deletion / Expiry
    Soft delete: mark expired=true, cron job purges rows nightly
    Redis TTL matches DB expiry: expired keys naturally evict from cache

---
Common Mistake

  Using sequential integer IDs as the short code (e.g. /1, /2, /3).
  → enumerable: anyone can scrape all URLs by iterating
  → reveals business volume (URL count = transaction count)
  → Fix: always encode IDs in base62 with at least 6-7 characters

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
2. News Feed (Twitter / Facebook Timeline)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Core insight: fan-out on write is cheap for normal users but catastrophically expensive for
              celebrities; the correct answer is always a hybrid.

---
Clarify First

  - Follow model or friend model? (directed graph vs undirected → different fan-out math)
  - What goes in the feed? (posts only, or likes/comments/shares too?)
  - Ordering: chronological or ranked?
  - Max followers a single user can have?
  - Feed length limit? (infinite scroll vs fixed N)

---
Scale Estimation

  500M users, 200M DAU
  Average follows: 200 → each post fans out to 200 feeds
  100M posts/day → ~1,200 writes/sec
  200M users × 20 feed reads/day → ~46,000 reads/sec
  Feed cache per user: 200 items × 8 bytes → 1.6KB per user → 800GB for all active users

---
Fan-out on Write (Push Model)

  On post: writer's post → async job → push postID into every follower's feed cache
  Feed cache: Redis sorted set, score = timestamp, trimmed to 1,000 items

  Flow:
    User A posts → message queue → fan-out workers → ZADD feed:followerB postID timestamp
                                                    → ZADD feed:followerC postID timestamp
                                                    → ...

  Reads: ZRANGE feed:userID 0 19 → 20 most recent postIDs → batch fetch post objects from DB

  Tradeoffs:
    + reads are O(1) — feed already assembled
    - writes are O(followers) — 10M followers = 10M Redis writes per post
    - celebrity with 100M followers: fan-out takes minutes
    - storage: every post stored N times (once per follower)

---
Fan-out on Read (Pull Model)

  On read: fetch list of followees → for each, fetch their last N posts → merge + sort

  Flow:
    User reads feed → get followee IDs → scatter: fetch posts:userID for each followee
                   → gather all results → merge sort by timestamp → return top 20

  Tradeoffs:
    + writes are O(1) — just store the post once
    - reads are O(followees) — slow; 1,000 followees = 1,000 DB lookups per page load
    - not cacheable effectively — every user's merge is unique
    - bad at scale: 46K reads/sec × 1,000 followees = 46M DB reads/sec

---
Hybrid Model (Production Standard)

  Threshold: if follower count < 1M → fan-out on write (normal users)
             if follower count ≥ 1M → fan-out on read (celebrities)

  Read path:
    1. Load user's pre-assembled feed cache (fan-out-on-write users they follow)
    2. Identify celebrities in followee list
    3. Fetch celebrity posts directly and merge in
    4. Return merged, sorted feed

  Why 1M threshold: fan-out to 1M is expensive but tractable; 10M+ is unacceptable latency.

---
Storage Schema

  posts table:
  ┌────────────┬──────────┬──────────────┬────────────┐
  │ post_id    │ user_id  │ content      │ created_at │
  │ BIGINT PK  │ BIGINT   │ TEXT         │ TIMESTAMP  │
  └────────────┴──────────┴──────────────┴────────────┘

  feed_cache (Redis sorted set per user):
    Key:   feed:{user_id}
    Value: post_id
    Score: unix timestamp (for ordering)
    Trim:  keep top 1,000 entries (ZREMRANGEBYRANK)

  follows table:
  ┌──────────────┬──────────────┐
  │ follower_id  │ followee_id  │
  │ BIGINT       │ BIGINT       │
  └──────────────┴──────────────┘
  Composite PK on (follower_id, followee_id), index on followee_id for reverse lookup.

---
Common Mistake

  Designing fan-out on write for all users without addressing the celebrity / hotspot problem.
  → a single Katy Perry post (100M followers) would saturate write workers for minutes
  → interviewer expects you to raise this unprompted and propose the hybrid

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
3. Chat System (WhatsApp / Messenger)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Core insight: messages must be ordered within a conversation, delivered exactly once, and reach
              offline users reliably — three independent hard problems.

---
Clarify First

  - 1-on-1 only, or group chat?
  - Max group size?
  - Message types: text only, or media (images, video)?
  - Delivery guarantees: at-least-once, at-most-once, exactly-once?
  - Read receipts and online presence required?

---
Scale Estimation

  1B users, 500M DAU
  50 messages/user/day → 25B messages/day → ~290,000 messages/sec
  Message size: ~100 bytes avg → 2.5TB messages/day stored
  Connection cost: WebSocket per active user → 500M concurrent connections → need connection servers (stateful)

---
Transport — Why WebSockets

  HTTP polling: client requests every N seconds → high latency (up to N seconds delay), wasteful
  Long polling: server holds request open until message arrives → better, but reconnect overhead
  WebSocket: persistent bidirectional TCP connection, full-duplex
    + server pushes messages instantly
    + one persistent connection per client
    - stateful: connection pinned to a specific connection server
    - harder to load balance (can't use stateless round-robin)

  Architecture:
    Client ←WebSocket→ Connection Server (stateful) ←→ Message Service ←→ DB

---
Message Ordering — Sequence Numbers

  Problem: if two clients send simultaneously, messages can arrive at server in different order.

  Solution: assign sequence numbers within each conversation (not globally)
    - Conversation has a monotonically increasing sequence_id
    - Message service atomically increments sequence_id per conversation on each write
    - Client renders messages sorted by sequence_id, not arrival order
    - Gaps in sequence detected client-side → client requests missing range from server

  Storage schema:
  ┌───────────────┬───────────────┬──────────────┬──────────┬────────────┬────────────┐
  │ conversation  │ sequence_id   │ sender_id    │ content  │ created_at │ msg_type   │
  │ BIGINT        │ BIGINT        │ BIGINT       │ TEXT     │ TIMESTAMP  │ ENUM       │
  └───────────────┴───────────────┴──────────────┴──────────┴────────────┴────────────┘
  PK: (conversation_id, sequence_id) — enables efficient range scans for history
  Shard key: conversation_id — all messages of a conversation on same shard

---
Offline Message Delivery

  Problem: recipient is offline when message arrives → must be delivered when they reconnect.

  Inbox table (undelivered messages):
  ┌──────────────┬────────────┬───────────────┬──────────────────┐
  │ recipient_id │ message_id │ conversation  │ created_at       │
  └──────────────┴────────────┴───────────────┴──────────────────┘

  Flow:
    Message arrives → stored in messages table + inserted into recipient's inbox
    Recipient reconnects → fetch all inbox rows → deliver + delete inbox rows
    Message marked delivered → send delivery receipt to sender

  Push notifications (device offline):
    Connection server detects user offline → push to APNs (iOS) / FCM (Android)
    Push carries message preview only; full message fetched on app open
    Push is best-effort — not guaranteed delivery

---
Online Presence

  Mechanism: client sends heartbeat every 5 seconds over WebSocket connection
  Connection server: updates last_seen timestamp in Redis (TTL = 10s)
  Any server can query Redis to check presence: GET presence:{user_id}

  Tradeoff: presence is eventually consistent — max 10s lag before detecting offline
  Do not store presence in a database — write rate is too high (N users × 1 heartbeat/5s)

---
Common Mistake

  Not addressing offline delivery and push notifications.
  → interviewer expects: what happens if the recipient is offline?
  → must describe: inbox table, reconnect flush, push notification fallback

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
4. Notification System
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Core insight: notifications are a fan-out problem at high volume with multiple delivery channels,
              each with different reliability characteristics.

---
Clarify First

  - Channels required: push (mobile), SMS, email?
  - Delivery guarantee: at-least-once or exactly-once?
  - User preferences: opt-out per channel, per notification type?
  - Latency target: real-time (<1s) or batched (hourly digest)?
  - Volume: events/sec triggering notifications?

---
Scale Estimation

  1B users
  10M notification events/day from product services → ~115 events/sec input
  Each event fans out to N subscribers → 1B notifications/day → ~11,600 notifications/sec
  Push: instant. SMS: ~1s. Email: ~30s. Delivery rate varies by channel.

---
Architecture — Async Fan-out Pipeline

  Triggering service (payment, social, marketing)
       ↓  event published
  Notification Queue (Kafka topic per event type)
       ↓  consumed by
  Notification Workers
       ↓  look up user preferences + channels
  Per-channel queues: [Push Queue] [SMS Queue] [Email Queue]
       ↓
  Channel handlers: APNs/FCM  |  Twilio  |  SendGrid
       ↓
  Delivery log + retry queue

  Why Kafka input queue: decouples triggering service from notification latency.
  Triggering service never blocks waiting for notifications to send.

---
Delivery Channels

  Push notifications (APNs / FCM)
    - APNs for iOS, FCM for Android
    - Device token required (stored per user+device in DB)
    - Best-effort: no guarantee of delivery if device offline for too long (FCM stores 4 weeks)
    - Retry: exponential backoff on provider failure
    - Token expiry: handle InvalidRegistration errors → delete stale tokens

  SMS (Twilio, AWS SNS)
    - Higher delivery rate than push (works on dumbphones, no app required)
    - Higher cost per message (~$0.01 - $0.10)
    - Character limits (160 chars per segment, concatenation for longer)
    - Retry: Twilio webhooks for delivery status; retry on failure

  Email (SendGrid, AWS SES)
    - Lowest cost at scale
    - Lowest engagement / open rate
    - Bounces, spam filters complicate delivery
    - Bulk sending requires domain reputation management

---
Idempotency Keys

  Problem: if a worker crashes after sending but before marking delivered, it retries → duplicate send.

  Solution: each notification has an idempotency_key = hash(user_id + event_id + channel)
    - Stored in Redis with TTL = 24 hours
    - Before sending: check Redis → key exists → skip (already sent)
    - After sending: write key to Redis → safe to retry

  APNs and FCM support apns-collapse-id / collapse_key:
    - Multiple notifications with same collapse key → only latest delivered
    - Natural deduplication at the platform level for some use cases

---
Retry Strategy

  Exponential backoff: retry after 1s, 2s, 4s, 8s, 16s → max 5 retries → dead letter queue
  Dead letter queue: manual inspection for systematic failures (wrong token, bad phone number)
  Circuit breaker: if >10% of sends to a provider fail → stop sending → alert on-call → fallback channel

---
User Preferences

  preferences table:
  ┌──────────┬──────────────────┬─────────────────┬──────────┐
  │ user_id  │ notification_type│ channel         │ enabled  │
  │ BIGINT   │ VARCHAR          │ ENUM(push,sms,  │ BOOL     │
  │          │                  │ email)          │          │
  └──────────┴──────────────────┴─────────────────┴──────────┘
  Worker checks preferences before dispatching to channel queues.
  Cache in Redis (TTL 5 min) — preferences rarely change, not worth a DB hit per notification.

---
Common Mistake

  Sending notifications synchronously within the triggering service request.
  → notification send latency (esp. SMS/email) blocks the main request
  → provider outage cascades into main service unavailability
  → Fix: always publish to a queue; decouple notification sending entirely from the triggering path

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
5. Web Crawler
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Core insight: a crawler is a distributed BFS; the hard parts are deduplication, politeness,
              and priority — not the fetching itself.

---
Clarify First

  - Scope: full web (billions of pages) or a specific domain?
  - Freshness: one-time crawl or continuous recrawl?
  - Content types: HTML only, or also PDFs, images?
  - robots.txt compliance required?
  - Politeness constraints (max requests/sec per domain)?

---
Scale Estimation

  1B pages to crawl, 1 month deadline
  1B / (30 days × 86,400s) → ~400 pages/sec
  Avg page size: 500KB → 400 × 500KB = 200MB/sec storage write rate
  Storage: 1B × 500KB = 500TB total
  → need distributed crawl workers, not a single machine

---
High-Level Flow

  URL Frontier (priority queue)
       ↓  dequeue URL
  Fetcher Workers (HTTP GET)
       ↓  raw HTML
  Parser (extract links + content)
    ↓                      ↓
  Link Extractor        Content Processor
  (normalize URLs)      (store to S3 / index)
       ↓
  Dedup Filter
       ↓ (unseen only)
  URL Frontier (enqueue new URLs)

---
BFS vs DFS

  DFS: follow links depth-first
    - Risk: gets trapped in deep link chains (infinite pagination, generated URLs)
    - Risk: one domain monopolizes the entire crawl
    - Not used in practice for general crawlers

  BFS: process URLs level by level from seeds
    + covers the web broadly before going deep
    + naturally distributes across many domains
    + important pages (many inbound links) reached sooner
    Used by: all major web crawlers (Googlebot, Bingbot)

---
URL Frontier — Priority Queue Design

  Not a simple queue — URLs need priority based on:
    - Page importance (PageRank estimate, number of inbound links)
    - Recency (recently updated pages crawled more often)
    - Politeness (max N requests/sec per domain)

  Two-tier frontier:
    Front queues: N priority queues, each holding URLs of a priority level
    Back queues: one queue per domain (ensures politeness — one domain per queue)

  Scheduler maps from front queues → back queues:
    - Picks a back queue to dequeue from (respects per-domain crawl delay)
    - FIFO within each domain queue (fairness)

---
Deduplication — Bloom Filter

  Problem: at 1B URLs, checking DB for every URL = 400 DB lookups/sec just for dedup

  Bloom filter: probabilistic set membership data structure
    - Space: ~10 bits per element → 1B URLs = 10Gb = 1.25GB in memory
    - Lookup: O(k) hash functions, no disk I/O
    - False positives: possible (filter says "seen" but URL is new → URL skipped)
    - False negatives: impossible (filter says "unseen" → URL is definitely new)

  Tradeoff: small fraction of new URLs skipped due to false positives → acceptable for crawling
  Alternative for dedup: MD5/SHA256 hash of normalized URL stored in a DB or Redis set
    → exact dedup but requires DB round trip per URL

---
Politeness — Robots.txt

  robots.txt: file at domain root specifying which paths crawlers may access
    Example: User-agent: * Disallow: /private/

  Crawler must:
    - Fetch robots.txt for each domain before crawling any URL in that domain
    - Cache robots.txt per domain (TTL: hours, not per-request)
    - Respect Crawl-delay directive
    - Never crawl disallowed paths

  Rate limiting per domain:
    - Track last_crawl_time per domain in Redis
    - Enforce minimum gap (e.g. 1 request/sec/domain by default)
    - Violations risk IP ban from the target domain

---
DNS Caching

  Problem: DNS lookup per URL = 10-100ms latency + hammers DNS servers
  Solution: local DNS cache per crawl worker, TTL = 30 minutes
  Benefit: eliminates DNS latency from the hot path, reduces DNS queries by 99%+

---
Common Mistake

  Not addressing politeness — designing a crawler that hammers a single domain
  without rate limiting per domain.
  → interviewer reads this as: you've never thought about the operational impact of crawling
  → fix: raise politeness (per-domain rate limiting + robots.txt) unprompted

