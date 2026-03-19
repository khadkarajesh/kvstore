# Engineering Blog Reading List

## How to read each post (the habit)

Before reading, answer the three pre-reading questions from memory.
While reading, find: what problem → what they tried → why it failed → what they built.
After reading, write 3 sentences: the problem, the solution, the tradeoff they accepted.

If you cannot write those 3 sentences after reading, read it again.
Store your 3-sentence summaries in: interview/reading-notes.md

---

## Reading Order — by interview frequency

Reordered by how often each topic appears in real system design interviews.
Read week 1 before week 2. Each week builds on the previous.

Priority legend:
  🔴 Critical — appears in almost every interview
  🟡 High     — appears in most interviews
  🟢 Good     — appears in specialist roles or senior interviews

---

### Week 1 — 🔴 Consistent Hashing (algorithm used everywhere)

**Tom White — Consistent Hashing**
URL: https://tom-e-white.com/2007/11/consistent-hashing.html
Topic: Consistent hashing algorithm, virtual nodes, data redistribution

Why first:
  Consistent hashing is referenced in almost every post that follows —
  Slack uses it for WebSocket routing, Cloudflare for rate limiting,
  Cassandra for data distribution, Redis Cluster for sharding.
  Read this once and every other post becomes clearer.

Pre-reading questions:
  1. You have 4 cache servers. You use hash(key) % 4 to route requests.
     You add a 5th server. What fraction of your cache keys now route
     to a different server?
  2. Why is that a problem for a caching layer?
  3. What does "virtual node" mean and how does it improve load distribution?

What to find while reading:
  - Why does modular hashing (% N) fail when N changes?
  - How does consistent hashing limit redistribution to only 1/N of keys?
  - What is a virtual node and how does it solve uneven load distribution?
  - When a node is added or removed, exactly which keys move and where?

Real-world uses (know these cold):
  Cassandra     → data distribution across nodes
  Redis Cluster → slot assignment across shards
  Kafka         → partition assignment to brokers
  Load balancers → routing sticky sessions
  CDN           → routing requests to edge nodes
  Slack         → routing WebSocket connections to Channel Servers

---

### Week 2 — 🔴 Caching at Scale (in every design)

**Facebook — Scaling Memcache at Facebook**
URL: https://engineering.fb.com/2013/04/15/core-infra/scaling-memcache-at-facebook/
Topic: Caching at extreme scale, thundering herd, cross-region consistency

Why second:
  Caching appears in every system design. Before you can design Twitter,
  WhatsApp, or YouTube, you need to understand cache invalidation,
  thundering herd, and cross-region consistency. This post is the
  most thorough treatment of all three.

Pre-reading questions:
  1. Facebook reads data 100x more than it writes. How does caching help,
     and what problem does it create when data changes?
  2. What is a thundering herd on cache miss? When does it happen?
  3. You have a cache cluster in the US and one in Europe. A user updates
     their profile. How do you keep both caches consistent?

What to find while reading:
  - What is lease-based cache invalidation and what problem does it solve?
  - How does Facebook handle the thundering herd (many simultaneous cache misses)?
  - What is the regional consistency problem and how do they solve it?
  - What is mcrouter and what does it add on top of standard memcached?

---

### Week 3 — 🔴 Feed Systems (most asked design question)

**Twitter — Timelines at Scale**
URL: https://www.infoq.com/presentations/Twitter-Timeline-Scalability/
Topic: Fan-out on write vs fan-out on read, the celebrity problem

Why third:
  "Design Twitter/Instagram/Facebook feed" is the single most common
  system design interview question. The fan-out on write vs read
  decision and the celebrity problem are asked at almost every company.
  This post gives you the real answer from the people who built it.

Pre-reading questions:
  1. When a user posts a tweet, you could write it to all followers' feeds
     immediately OR compute each user's feed at read time.
     What are the tradeoffs of each?
  2. A celebrity has 50M followers. They tweet. What breaks with
     fan-out on write?
  3. How do you merge two feed sources at read time without the
     user noticing latency?

What to find while reading:
  - What is fan-out on write (push) and what breaks it for celebrities?
  - How does Twitter's hybrid model work — who gets pushed, who gets pulled?
  - What is the follower threshold for switching between push and pull?
  - How does Redis store each user's pre-computed timeline?

---

### Week 4 — 🔴 Kafka Internals (used in every scalable design)

**LinkedIn — Open-sourcing Kafka**
URL: https://blog.linkedin.com/2011/01/11/open-source-linkedin-kafka
Topic: Why Kafka was built, what existing tools failed at

Why fourth:
  Kafka appears in almost every system design beyond a simple CRUD app.
  Understanding WHY it was built (what ActiveMQ/RabbitMQ couldn't do)
  gives you the ability to justify using it — not just name-drop it.

Pre-reading questions:
  1. LinkedIn needs to move activity events from web servers to Hadoop.
     What existing tools would you consider?
  2. ActiveMQ and RabbitMQ exist. Why would LinkedIn build something new?
  3. What does "durable" mean for a message queue and why does it matter?

What to find while reading:
  - What specific problems did LinkedIn have with existing message queues?
  - How does Kafka's design (sequential writes, consumer pull) differ?
  - What guarantee does Kafka make that traditional queues don't?
  - What did LinkedIn give up to get Kafka's throughput?

---

### Week 5 — 🔴 Real-Time WebSocket Fan-out

**Uber — Uber's Real-Time Push Platform (RAMEN)**
URL: https://eng.uber.com/real-time-push-platform/
Topic: Pushing events to millions of mobile clients in real time

Pre-reading questions:
  1. You need to push driver location updates to riders every 2 seconds.
     What delivery mechanism do you use?
  2. Mobile clients in developing markets have unstable connections.
     How does that change your architecture?
  3. How do you route a location update to the specific server holding
     a rider's connection?

What to find while reading:
  - What does RAMEN stand for and what problem does it solve?
  - How does Uber route events to the correct connection server?
  - How do they handle bandwidth-constrained mobile clients?
  - What happens to events when a client is temporarily offline?

---

**Slack — Real-time Messaging Architecture**
URL: https://slack.engineering/real-time-messaging/
Topic: WebSocket fan-out to channel members, Channel Server routing

Pre-reading questions:
  1. A Slack channel has 5,000 members. One sends a message. How do you
     deliver it to all 5,000 across 200 different WebSocket servers?
  2. What is consistent hashing and how would you apply it here?
     (You read Tom White in week 1 — apply it now)
  3. What happens when a Channel Server goes down?

What to find while reading:
  - What are Channel Servers and what problem do they solve?
  - How does Slack use consistent hashing for message routing?
  - What happens when a Channel Server goes down?
  - How do they handle members who are offline?

---

### Week 6 — 🔴 Database Sharding + ID Generation

**Pinterest — Sharding Pinterest: How we scaled our MySQL fleet**
URL: https://medium.com/pinterest-engineering/sharding-pinterest-how-we-scaled-our-mysql-fleet-3f341e96ca6f
Topic: Static sharding design, 64-bit ID encoding shard + type

Pre-reading questions:
  1. Your MySQL single instance is running out of capacity. What options exist?
  2. What is the difference between static sharding and consistent hashing?
  3. If you shard a table, how does a query needing data from multiple shards work?

What to find while reading:
  - Why did Pinterest choose static sharding over consistent hashing?
  - How does their 64-bit ID encode shard + type + local ID?
  - What operational advantage does static sharding give?
  - What can they never do easily because of static sharding?

---

**Twitter — Announcing Snowflake**
URL: https://blog.x.com/engineering/en_us/a/2010/announcing-snowflake
Topic: Distributed unique ID generation, time-ordered IDs

Pre-reading questions:
  1. You need globally unique IDs for tweets. Why is UUID a bad choice?
  2. What happens if two machines generate IDs at the exact same millisecond?
  3. Why does ID sort order matter for a Cassandra clustering key?

What to find while reading:
  - What problem did Twitter have with their previous ID scheme?
  - How does Snowflake pack timestamp + machine + sequence into 64 bits?
  - What is the clock skew problem and how does Snowflake handle it?

---

**Instagram — Sharding & IDs at Instagram**
URL: https://instagram-engineering.com/sharding-ids-at-instagram-1cf5a71e5a5c
Topic: ID generation inside sharded Postgres without a central coordinator

Pre-reading questions:
  1. Twitter uses a central service for IDs. What is the downside?
  2. You have 1,000 Postgres shards. How do you generate unique IDs without
     them talking to each other?

What to find while reading:
  - Why did Instagram avoid a central ID service?
  - How does their ID scheme encode the shard ID?
  - What tradeoff did they accept compared to Snowflake?

---

### Week 7 — 🔴 Message Storage + Fault Tolerance

**Discord — How Discord Stores Billions of Messages**
URL: https://discord.com/blog/how-discord-stores-billions-of-messages
Topic: Cassandra data modeling for chat, partition key design

Pre-reading questions:
  1. You are storing WhatsApp messages. What is your partition key and why?
  2. What happens to Cassandra read performance if your partition grows unboundedly?
  3. Why is Cassandra preferred over PostgreSQL for chat message storage?

What to find while reading:
  - Why did Discord choose Cassandra over their previous solution?
  - What specific Cassandra data modeling problem did they hit?
  - How did they solve unbounded partition growth?

---

**Netflix — Fault Tolerance in a High Volume, Distributed System**
URL: https://netflixtechblog.com/fault-tolerance-in-a-high-volume-distributed-system-91ab4faae74a
Topic: Circuit breakers, bulkheads, fallbacks (Hystrix)

Pre-reading questions:
  1. One Netflix API call fans out to 6 downstream services. One is slow (3s).
     What happens to your thread pool?
  2. What is a circuit breaker? How is it different from a retry?
  3. What is a fallback in the context of a failed service call?

What to find while reading:
  - What failure cascade did Netflix observe before Hystrix?
  - What is a bulkhead and why does it prevent cascading failure?
  - When does a circuit breaker open and when does it close again?
  - What fallback strategy does Netflix use when a service is unavailable?

---

### Week 8 — 🔴 Notification Systems + Typeahead

**Notification systems — Uber RAMEN (second read)**
URL: https://eng.uber.com/real-time-push-platform/
Re-read with different focus: APNs/FCM delivery to offline/mobile devices.

First read (Week 5) focus: WebSocket fan-out to online clients.
Second read (Week 8) focus: push notification delivery to offline devices.

Pre-reading questions for second read:
  1. A user's phone is locked. Your app needs to notify them of a new message.
     How does the notification reach the device?
  2. APNs/FCM guarantee at-most-once, at-least-once, or exactly-once?
  3. You send 10M push notifications simultaneously. How do you prevent
     APNs/FCM from rate-limiting you?

What to find while reading (second pass):
  - How does Uber route a notification to the correct APNs/FCM endpoint?
  - What happens when a device is offline for 3 days?
  - How do they handle 70,000 pushes/second without overwhelming APNs/FCM?
  - Data message vs display notification — what is the difference?

Notification design patterns for interviews:
  Online user:    deliver via WebSocket directly
  Offline user:   store in DB + APNs/FCM → device wakes → fetches from DB
  Fan-out:        Kafka → Notification Service → APNs/FCM (parallel, batched)
  Deduplication:  notification_id deduplicates on device
  Retry:          APNs/FCM has own retry — don't double-retry from your service

---

**LinkedIn — Cleo: Technology Behind LinkedIn's Typeahead Search**
URL: https://engineering.linkedin.com/open-source/cleo-open-source-technology-behind-linkedins-typeahead-search
Topic: Autocomplete at scale, prefix indexing, ranking

Pre-reading questions:
  1. A user types "mar" in the search box. You need results in under 50ms.
     How would you index names for prefix search?
  2. What is a trie data structure and how does it support prefix lookup?
  3. You have 300M users. Storing a full trie in memory is expensive.
     How do you shard typeahead across machines?

What to find while reading:
  - What data structure does Cleo use for prefix matching?
  - How does it handle out-of-order terms ("john smi" matching "John Smith")?
  - How does LinkedIn rank results — why is CEO ranked above intern?
  - How is the index updated when a new user joins?

Typeahead design patterns for interviews:
  Simple:   trie in memory, single server, O(prefix length) lookup
  At scale: shard trie by first letter, replicate for read throughput
  Ranking:  score = frequency × recency × personalization weight
  Updates:  eventual consistency — index updates async

---

### Week 9 — 🟡 Object Storage + Video Streaming

**Facebook — Needle in a Haystack: Efficient Storage of Billions of Photos**
URL: https://engineering.fb.com/2009/04/30/core-infra/needle-in-a-haystack-efficient-storage-of-billions-of-photos/
Topic: Object storage, aggregating small files, avoiding metadata bottleneck

Pre-reading questions:
  1. You store 1 billion photos. What is the bottleneck when fetching a photo
     from a standard POSIX filesystem?
  2. Why does reading a small file from disk require multiple I/O operations?
  3. What is the difference between object storage and a filesystem?

What to find while reading:
  - What specific POSIX filesystem problem did Facebook hit at scale?
  - How does Haystack aggregate millions of photos into one large file?
  - What is a "needle" and what is a "haystack"?
  - How does the in-memory index avoid disk I/O on every read?

---

**Netflix — Per-Title Encode Optimization**
URL: https://netflixtechblog.com/per-title-encode-optimization-7e99442b62a2
Topic: Video encoding pipeline, adaptive bitrate streaming

Pre-reading questions:
  1. A user on 3G switches to WiFi. How does Netflix adjust quality mid-stream
     without stopping playback?
  2. What is a bitrate ladder and why does one-size-fits-all waste bandwidth?
  3. Should a cartoon and an action movie use the same encoding settings? Why?

What to find while reading:
  - What is adaptive bitrate streaming (ABR)?
  - What is a bitrate ladder and what does per-title optimization solve?
  - How does Netflix measure visual complexity of a title?
  - What is the CDN's role — what does it cache?

Video delivery architecture (know this separately — not in the post):
  Client → CDN edge → origin (Netflix Open Connect)
  File split into 2-10 second chunks, each at multiple bitrates.
  Client monitors bandwidth → requests next chunk at right bitrate.
  Small chunks: faster quality switch, more requests.
  Large chunks: fewer requests, slower quality adaptation.

---

### Week 10 — 🟡 Graph Storage + Search

**Facebook — TAO: The Power of the Graph**
URL: https://engineering.fb.com/2013/06/25/core-infra/tao-the-power-of-the-graph/
Topic: Distributed graph storage, objects + associations model

Pre-reading questions:
  1. Facebook's social graph has billions of edges. Why is a relational DB a poor fit?
  2. Most social graph operations are reads. How does this change your storage design?
  3. What is the difference between an object and an association in a graph model?

What to find while reading:
  - What is TAO and what two data types does it model?
  - Why did Facebook build TAO instead of using Memcache directly?
  - How does TAO handle read-heavy workloads across multiple datacenters?
  - What consistency model does TAO provide and what does it sacrifice?

---

**Airbnb — Elasticsearch at Airbnb**
URL: https://medium.com/airbnb-engineering/tagged/elasticsearch
Topic: Search indexing, denormalization, real-time index updates

Pre-reading questions:
  1. What is an inverted index and why is it faster than a full-text DB scan?
  2. Airbnb listings have 50 fields. How do you index this for fast multi-field search?
  3. A listing's price changes. How do you update the search index without downtime?

What to find while reading:
  - Why did Airbnb choose Elasticsearch over a relational DB for search?
  - How do they keep the search index in sync with the primary DB?
  - What is denormalization in search and why is it necessary?
  - What happens to search results during an index update?

---

### Week 11 — 🟡 Observability + Distributed Locking

**Uber — Evolving Distributed Tracing at Uber Engineering**
URL: https://www.uber.com/blog/distributed-tracing/
Topic: Distributed tracing across microservices, Jaeger

Pre-reading questions:
  1. A request touches 20 microservices. One is slow. How do you find which one?
  2. What is a trace and what is a span?
  3. At 1M requests/sec, storing every trace is expensive. How do you sample?

What to find while reading:
  - What specific debugging problem prompted Uber to build Jaeger?
  - What is the difference between a trace and a span?
  - Why did Uber switch from pull-based to push-based trace collection?
  - Head-based vs tail-based sampling — what tradeoff does each make?

Interview signal — mention these unprompted in every design:
  Metrics: "trip completion rate dropped 20% in NYC" (dashboard)
  Traces:  "which service added 500ms to this request?" (debugging)
  Logs:    "what exactly happened for order #42?" (investigation)

---

**Martin Kleppmann — How to do distributed locking**
URL: https://martin.kleppmann.com/2016/02/08/how-to-do-distributed-locking.html
Topic: When Redis locks fail, fencing tokens, Redlock critique

Pre-reading questions:
  1. Process A acquires a Redis lock (TTL 5s), pauses for 6s (GC pause).
     TTL expires. Process B acquires the same lock. Process A resumes.
     What happens?
  2. What is a fencing token and how does it prevent this?
  3. What assumptions does Redlock make about clocks?

What to find while reading:
  - What specific scenario breaks Redlock?
  - What is a fencing token and how does it provide safety TTLs cannot?
  - When is Redis SETNX actually sufficient?
  - What does he recommend for correctness-critical locks?

---

### Week 12 — 🟡 Rate Limiting + Geo-indexing

**Cloudflare — How we built rate limiting capable of scaling to millions of domains**
URL: https://blog.cloudflare.com/counting-things-a-lot-of-different-things/
Topic: Distributed rate limiting without shared global state

Pre-reading questions:
  1. You need to rate limit a user to 100 req/min across 1,000 edge servers.
     How do you count without a central counter?
  2. What is a leaky bucket? How is it different from a token bucket?
  3. What is the tradeoff between accuracy and performance in distributed rate limiting?

What to find while reading:
  - Why is a central Redis counter not viable at CDN edge scale?
  - What approximation technique does Cloudflare use?
  - How much accuracy do they sacrifice and why is that acceptable?
  - What is the role of consistent hashing in their rate limiter?

---

**Uber — H3: Hexagonal Hierarchical Spatial Index**
URL: https://www.uber.com/en-US/blog/h3/
Topic: Geospatial indexing for dispatch, surge pricing

Pre-reading questions:
  1. You need to find all drivers within 2km of a rider. How do you query
     a database for this efficiently?
  2. Why are square grid cells problematic for distance calculations?
  3. What is a spatial index?

What to find while reading:
  - Why did Uber choose hexagons over squares or geographic boundaries?
  - What does "hierarchical" mean in H3?
  - How does H3 enable surge pricing calculations across a city?

---

### Week 13 — 🟢 Operational Depth

**Stripe — Online Migrations at Scale**
URL: https://stripe.com/blog/online-migrations
Topic: Live database schema migration without downtime

Pre-reading questions:
  1. You need to rename a column in a table with 500M rows. How without downtime?
  2. What is dual-write and when do you use it?
  3. What can go wrong if you flip reads before backfilling is complete?

What to find while reading:
  - What are the four phases of Stripe's migration pattern?
  - Why write to both old and new tables simultaneously?
  - How do they verify backfill is complete before flipping reads?
  - What is the rollback plan if something goes wrong after the flip?

---

**Dropbox — Rewriting the heart of our sync engine**
URL: https://dropbox.tech/infrastructure/rewriting-the-heart-of-our-sync-engine
Topic: When to rewrite vs fix, deterministic event-loop architecture

Pre-reading questions:
  1. Dropbox syncs files across millions of devices. What are the consistency requirements?
  2. Why is multithreaded code hard to reason about for correctness?
  3. What does "deterministic" mean in the context of a sync engine?

What to find while reading:
  - What specific correctness problems did the old engine have?
  - Why did they choose Rust for the rewrite?
  - What is the event-loop architecture and how does it eliminate threading bugs?
  - What did they give up (performance, features) to get determinism?

---

### Week 14 — 🟢 Deep Storage Evolution

**Discord — How Discord Stores Trillions of Messages**
URL: https://discord.com/blog/how-discord-stores-trillions-of-messages
Topic: When Cassandra breaks at scale — ScyllaDB migration

Pre-reading questions:
  1. Cassandra is JVM-based. What GC problem does that cause at high throughput?
  2. You have 177 Cassandra nodes. What would make you consider migrating away?
  3. What is the risk of migrating a live message store with no downtime?

What to find while reading:
  - What specific metrics were unacceptable in Cassandra?
  - Why ScyllaDB instead of fixing Cassandra?
  - How did they migrate without losing messages or going offline?
  - What did they build in Rust and why?

---

## Already Read (from distributed transactions study session)

- Airbnb — Avoiding Double Payments (Orpheus idempotency library)
  https://medium.com/airbnb-engineering/avoiding-double-payments-in-a-distributed-payments-system-2981f6b070bb

- Airbnb — Measuring Transactional Integrity
  https://medium.com/airbnb-engineering/measuring-transactional-integrity-in-airbnbs-distributed-payment-ecosystem-a670d6926d22

- Uber — Fulfillment Platform Re-architecture (saga → Spanner)
  https://eng.uber.com/fulfillment-platform-rearchitecture/

- Uber — Cadence open-source orchestration
  https://www.uber.com/blog/open-source-orchestration-tool-cadence-overview/

- DoorDash — Cadence as fallback for event-driven processing
  https://careersatdoordash.com/blog/building-reliable-workflows-cadence-as-a-fallback-for-event-driven-processing/

- Shopify — Capturing every change from sharded monolith (Debezium CDC)
  https://shopify.engineering/capturing-every-change-shopify-sharded-monolith

- Netflix — WAL data platform
  https://netflixtechblog.com/building-a-resilient-data-platform-with-write-ahead-log-at-netflix-127b6712359a

- Stripe — Designing robust APIs with idempotency
  https://stripe.com/blog/idempotency

---

## Topic Coverage Map

  🔴 Critical (weeks 1-8):
    Consistent hashing, Caching, Feed fan-out, Kafka, Real-time WebSocket,
    Database sharding, ID generation, Message storage, Fault tolerance,
    Notification systems, Typeahead

  🟡 High (weeks 9-12):
    Object storage, Video streaming, Graph storage, Search indexing,
    Observability / tracing, Distributed locking, Rate limiting, Geo-indexing

  🟢 Good (weeks 13-14):
    Live database migrations, Deterministic architecture, Storage evolution

  ✅ Already covered (distributed transactions session):
    Saga pattern, Outbox pattern, 2PC, Idempotency, CDC, Compensation

---

## After Reading — Write 3 Sentences

For every post, write in your own words:
  1. The problem they had (specific — not "they needed to scale")
  2. What they built and the core mechanism
  3. The tradeoff they accepted

Store in: interview/reading-notes.md
