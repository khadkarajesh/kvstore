# Engineering Blog Reading List

## How to read each post (the habit)

Before reading, answer the three pre-reading questions from memory.
While reading, find: what problem → what they tried → why it failed → what they built.
After reading, write 3 sentences: the problem, the solution, the tradeoff they accepted.

If you cannot write those 3 sentences after reading, read it again.

---

## Reading Order

Ordered by foundational importance. Do not skip ahead.
One post per week minimum. Two if you have time.

---

### Week 1 — Storage fundamentals

**Discord — How Discord Stores Billions of Messages**
URL: https://discord.com/blog/how-discord-stores-billions-of-messages
Topic: Cassandra data modeling for chat

Pre-reading questions:
  1. You are storing WhatsApp messages. What is your partition key and why?
  2. What happens to Cassandra read performance if your partition grows unboundedly?
  3. Why is Cassandra preferred over PostgreSQL for chat message storage?

What to find while reading:
  - Why did Discord choose Cassandra over their previous solution?
  - What specific Cassandra data modeling problem did they hit?
  - How did they solve unbounded partition growth?

---

### Week 2 — Storage at extreme scale

**Discord — How Discord Stores Trillions of Messages**
URL: https://discord.com/blog/how-discord-stores-trillions-of-messages
Topic: When Cassandra breaks, what next (ScyllaDB migration)

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

### Week 3 — ID generation

**Twitter — Announcing Snowflake**
URL: https://blog.x.com/engineering/en_us/a/2010/announcing-snowflake
Topic: Distributed unique ID generation

Pre-reading questions:
  1. You need globally unique IDs for tweets. Why is UUID a bad choice?
  2. What happens if two machines generate IDs at the exact same millisecond?
  3. Why does ID sort order matter for a database clustering key?

What to find while reading:
  - What problem did Twitter have with their previous ID scheme?
  - How does Snowflake pack timestamp + machine + sequence into 64 bits?
  - What is the clock skew problem and how does Snowflake handle it?

---

**Instagram — Sharding & IDs at Instagram**
URL: https://instagram-engineering.com/sharding-ids-at-instagram-1cf5a71e5a5c
Topic: ID generation inside sharded Postgres (no central coordinator)

Pre-reading questions:
  1. Twitter uses a central service for IDs. What is the downside of that approach?
  2. You have 1,000 Postgres shards. How do you generate unique IDs without them talking to each other?
  3. What is PL/pgSQL and why would you put ID generation logic inside the database?

What to find while reading:
  - Why did Instagram avoid a central ID service?
  - How does their ID scheme encode the shard ID?
  - What tradeoff did they accept compared to Snowflake?

---

### Week 4 — Real-time delivery

**Uber — Uber's Real-Time Push Platform (RAMEN)**
URL: https://eng.uber.com/real-time-push-platform/
Topic: Pushing events to millions of mobile clients in real time

Pre-reading questions:
  1. You need to push driver location updates to riders every 2 seconds. What delivery mechanism do you use?
  2. Mobile clients in developing markets have unstable connections. How does that change your architecture?
  3. How do you route a location update to the specific server holding a rider's connection?

What to find while reading:
  - What does RAMEN stand for and what problem does it solve?
  - How does Uber route events to the correct connection server?
  - How do they handle bandwidth-constrained mobile clients?
  - What happens to events when a client is temporarily offline?

---

**Slack — Real-time Messaging Architecture**
URL: https://slack.engineering/real-time-messaging/
Topic: WebSocket fan-out to channel members

Pre-reading questions:
  1. A Slack channel has 5,000 members. One member sends a message. How do you deliver it to all 5,000?
  2. 5,000 members are connected to 200 different WebSocket servers. How does the sending server know which servers to notify?
  3. What is consistent hashing and how would you apply it to WebSocket routing?

What to find while reading:
  - What are Channel Servers and what problem do they solve?
  - How does Slack use consistent hashing for message routing?
  - What happens when a Channel Server goes down?
  - How do they handle members who are offline?

---

### Week 5 — Fault tolerance

**Netflix — Fault Tolerance in a High Volume, Distributed System**
URL: https://netflixtechblog.com/fault-tolerance-in-a-high-volume-distributed-system-91ab4faae74a
Topic: Circuit breakers, bulkheads, fallbacks (Hystrix)

Pre-reading questions:
  1. One Netflix API call fans out to 6 downstream services. One service is slow (takes 3s). What happens to your thread pool?
  2. What is a circuit breaker? How is it different from a retry?
  3. What is a fallback in the context of a failed service call?

What to find while reading:
  - What failure cascade did Netflix observe before Hystrix?
  - What is a bulkhead and why does it prevent cascading failure?
  - When does a circuit breaker open and when does it close again?
  - What fallback strategy does Netflix use when a service is unavailable?

---

### Week 6 — Sharding and migrations

**Pinterest — Sharding Pinterest: How we scaled our MySQL fleet**
URL: https://medium.com/pinterest-engineering/sharding-pinterest-how-we-scaled-our-mysql-fleet-3f341e96ca6f
Topic: Static sharding design, 64-bit ID encoding

Pre-reading questions:
  1. Your MySQL single instance is running out of capacity. What are your options?
  2. What is the difference between static sharding and dynamic sharding (consistent hashing)?
  3. If you shard a table, how does a query that needs data from multiple shards work?

What to find while reading:
  - Why did Pinterest choose static sharding over consistent hashing?
  - How does their 64-bit ID encode shard + type + local ID?
  - What operational advantage does static sharding give them?
  - What can they never do easily because of static sharding?

---

**Stripe — Online Migrations at Scale**
URL: https://stripe.com/blog/online-migrations
Topic: Migrating a live database schema without downtime

Pre-reading questions:
  1. You need to rename a column in a table with 500M rows. How do you do it without downtime?
  2. What is dual-write and when do you use it?
  3. What can go wrong if you flip reads to a new table before backfilling is complete?

What to find while reading:
  - What are the four phases of Stripe's migration pattern?
  - Why do they write to both old and new tables simultaneously?
  - How do they verify the backfill is complete before flipping reads?
  - What is the rollback plan if something goes wrong after the flip?

---

### Week 7 — Rate limiting and graph storage

**Cloudflare — How we built rate limiting capable of scaling to millions of domains**
URL: https://blog.cloudflare.com/counting-things-a-lot-of-different-things/
Topic: Distributed rate limiting without shared global state

Pre-reading questions:
  1. You need to rate limit a user to 100 requests per minute across 1,000 edge servers. How do you count requests without a central counter?
  2. What is a leaky bucket algorithm? How is it different from a token bucket?
  3. What is the trade-off between accuracy and performance in distributed rate limiting?

What to find while reading:
  - Why is a central Redis counter not viable at CDN edge scale?
  - What approximation technique does Cloudflare use?
  - How much accuracy do they sacrifice and why is that acceptable?
  - What is the role of consistent hashing in their rate limiter?

---

**Facebook — TAO: The Power of the Graph**
URL: https://engineering.fb.com/2013/06/25/core-infra/tao-the-power-of-the-graph/
Topic: Distributed graph storage for the social graph

Pre-reading questions:
  1. Facebook's social graph has billions of edges (friendships, likes, comments). Why is a relational DB a poor fit?
  2. Most social graph operations are reads (check if A is friends with B, fetch A's friends). How does this change your storage design?
  3. What is the difference between an object and an association in a graph data model?

What to find while reading:
  - What is TAO and what two data types does it model?
  - Why did Facebook build TAO instead of using Memcache directly?
  - How does TAO handle read-heavy workloads across multiple datacenters?
  - What consistency model does TAO provide and what does it sacrifice?

---

### Week 8 — Message queues and sync engines

**LinkedIn — Open-sourcing Kafka**
URL: https://blog.linkedin.com/2011/01/11/open-source-linkedin-kafka
Topic: Why Kafka was built, what existing tools failed at

Pre-reading questions:
  1. LinkedIn needs to move activity events (pageviews, clicks) from web servers to Hadoop. What existing tools would you consider?
  2. ActiveMQ and RabbitMQ exist. Why would LinkedIn build something new?
  3. What does "durable" mean for a message queue and why does it matter for analytics pipelines?

What to find while reading:
  - What specific problems did LinkedIn have with existing message queues?
  - How does Kafka's design (sequential disk writes, consumer pull) differ from traditional queues?
  - What guarantee does Kafka make that traditional queues don't?
  - What did LinkedIn give up to get Kafka's throughput?

---

**Dropbox — Rewriting the heart of our sync engine**
URL: https://dropbox.tech/infrastructure/rewriting-the-heart-of-our-sync-engine
Topic: When to rewrite vs fix, deterministic event-loop architecture

Pre-reading questions:
  1. Dropbox syncs files across millions of devices. What are the consistency requirements?
  2. Why is multithreaded code hard to reason about for correctness?
  3. What does "deterministic" mean in the context of a sync engine?

What to find while reading:
  - What specific correctness problems did the old sync engine have?
  - Why did they choose Rust for the rewrite?
  - What is the event-loop architecture and how does it eliminate the threading bugs?
  - What did they have to give up (performance, features) to get determinism?

---

### Week 9 — Object storage and feed systems

**Facebook — Needle in a Haystack: Efficient Storage of Billions of Photos**
URL: https://engineering.fb.com/2009/04/30/core-infra/needle-in-a-haystack-efficient-storage-of-billions-of-photos/
Topic: Object storage — storing billions of files efficiently

Pre-reading questions:
  1. You store 1 billion photos. Each photo has metadata (name, permissions, path).
     What is the bottleneck when fetching a photo from a POSIX filesystem?
  2. Why does reading a small file from disk require multiple I/O operations?
  3. What is the difference between object storage and a filesystem?

What to find while reading:
  - What specific POSIX filesystem problem did Facebook hit at scale?
  - How does Haystack aggregate millions of photos into one file?
  - What is a "needle" and what is a "haystack" in their terminology?
  - How does the in-memory index avoid disk I/O on every read?

---

**Twitter — Timelines at Scale**
URL: https://www.infoq.com/presentations/Twitter-Timeline-Scalability/
Topic: Fan-out on write vs fan-out on read, the celebrity problem

Pre-reading questions:
  1. When a user posts a tweet, you could write it to all followers' feeds immediately
     OR compute each user's feed at read time. What are the tradeoffs of each?
  2. A celebrity has 50M followers. They tweet. What breaks with fan-out on write?
  3. How do you merge two feed sources at read time without the user noticing latency?

What to find while reading:
  - What is fan-out on write (push) and what breaks it for celebrities?
  - How does Twitter's hybrid model work — who gets pushed, who gets pulled?
  - What is the follower threshold for switching between push and pull?
  - How does Redis store each user's pre-computed timeline?

---

### Week 10 — Caching and search

**Facebook — Scaling Memcache at Facebook**
URL: https://engineering.fb.com/2013/04/15/core-infra/scaling-memcache-at-facebook/
Topic: Caching at extreme scale, consistency across regions

Pre-reading questions:
  1. Facebook reads data 100x more than it writes. How does caching help, and
     what problem does it create when data changes?
  2. What is a thundering herd on cache miss? When does it happen?
  3. You have a cache cluster in the US and one in Europe. A user updates their
     profile. How do you keep both caches consistent?

What to find while reading:
  - What is lease-based cache invalidation and what problem does it solve?
  - How does Facebook handle the thundering herd (many simultaneous cache misses)?
  - What is the regional consistency problem and how do they solve it?
  - What is mcrouter and what does it add on top of standard memcached?

---

**Airbnb — Elasticsearch at Airbnb**
URL: https://medium.com/airbnb-engineering/tagged/elasticsearch
Topic: Search indexing, denormalization for search, real-time index updates

Pre-reading questions:
  1. What is an inverted index and why is it faster than a database full-text scan?
  2. Airbnb listings have 50 fields (price, location, amenities, availability).
     How do you index this for fast multi-field search?
  3. A listing's price changes. How do you update the search index without
     taking the search service offline?

What to find while reading:
  - Why did Airbnb choose Elasticsearch over a relational DB for search?
  - How do they keep the search index in sync with the primary database?
  - What is denormalization in search and why is it necessary?
  - What happens to search results during an index update?

---

### Week 11 — Distributed locking and geo-indexing

**Martin Kleppmann — How to do distributed locking**
URL: https://martin.kleppmann.com/2016/02/08/how-to-do-distributed-locking.html
Topic: When Redis locks fail, fencing tokens, correctness vs performance

Pre-reading questions:
  1. You use Redis SETNX to implement a distributed lock. Process A acquires
     the lock, then pauses for 5 seconds (GC). The lock TTL expires. Process B
     acquires the same lock. Process A resumes. What happens?
  2. What is a fencing token and how does it prevent the above scenario?
  3. What assumptions does Redlock make about clocks that may not hold in practice?

What to find while reading:
  - What specific failure scenario does Kleppmann use to break Redlock?
  - What is a fencing token and how does it provide safety that TTLs cannot?
  - When is Redis SETNX actually sufficient for locking?
  - What does he recommend instead of Redlock for correctness-critical locks?

---

**Uber — H3: Hexagonal Hierarchical Spatial Index**
URL: https://www.uber.com/en-US/blog/h3/
Topic: Geospatial indexing for dispatch, surge pricing, and location queries

Pre-reading questions:
  1. You need to find all drivers within 2km of a rider. How do you query
     a database for this efficiently?
  2. Why are square grid cells problematic for distance calculations
     (hint: think about corners vs edges)?
  3. What is a spatial index and how is it different from a regular DB index?

What to find while reading:
  - Why did Uber choose hexagons over squares or geographic boundaries?
  - What does "hierarchical" mean in H3 — how do zoom levels work?
  - How does H3 enable surge pricing calculations across a city?
  - What problems does H3 solve that a simple lat/lng bounding box query cannot?

---

### Week 12 — Video streaming and consistent hashing

**Netflix — Per-Title Encode Optimization**
URL: https://netflixtechblog.com/per-title-encode-optimization-7e99442b62a2
Topic: Video encoding pipeline, adaptive bitrate streaming

Pre-reading questions:
  1. A user watches Netflix on a slow 3G connection, then switches to WiFi.
     How does Netflix adjust video quality without stopping playback?
  2. What is a bitrate ladder and why does one-size-fits-all waste bandwidth?
  3. A cartoon and an action movie have the same resolution. Should they
     use the same encoding settings? Why or why not?

What to find while reading:
  - What is adaptive bitrate streaming (ABR) and how does the client choose quality?
  - What is a bitrate ladder and what problem does per-title optimization solve?
  - How does Netflix measure visual complexity of a title?
  - What is the CDN's role in video delivery — what does it cache and why?

Additional context for interviews (not in this post — know this separately):
  Video delivery architecture:
    Client → CDN edge → origin (Netflix Open Connect)
    Video file split into 2-10 second chunks, each chunk encoded at multiple bitrates.
    Client player monitors bandwidth, requests next chunk at appropriate bitrate.
    Chunk size tradeoff: small chunks = faster quality switch, more requests.
                         large chunks = fewer requests, slower quality adaptation.

---

**Tom White — Consistent Hashing**
URL: https://tom-e-white.com/2007/11/consistent-hashing.html
Topic: Consistent hashing algorithm, virtual nodes, data redistribution

Pre-reading questions:
  1. You have 4 cache servers. You use hash(key) % 4 to route requests.
     You add a 5th server. What fraction of your cache keys now route to
     a different server?
  2. Why is that a problem for a caching layer?
  3. What does "virtual node" mean and how does it improve load distribution?

What to find while reading:
  - Why does modular hashing (% N) fail when N changes?
  - How does consistent hashing limit redistribution to only 1/N of keys?
  - What is a virtual node and how does it solve uneven load distribution?
  - When a node is added or removed, exactly which keys move and where?

Real-world applications of consistent hashing (know these for interviews):
  - Cassandra: data distribution across nodes
  - Redis Cluster: slot assignment across shards
  - Kafka: partition assignment to brokers
  - Load balancers: routing sticky sessions
  - CDN: routing requests to edge nodes

---

### Week 13 — Typeahead and observability

**LinkedIn — Cleo: Technology Behind LinkedIn's Typeahead Search**
URL: https://engineering.linkedin.com/open-source/cleo-open-source-technology-behind-linkedins-typeahead-search
Topic: Autocomplete / typeahead at scale, prefix indexing

Pre-reading questions:
  1. A user types "mar" in the search box. You need to return matching names
     in under 50ms. How would you index names for prefix search?
  2. What is a trie data structure and how does it support prefix lookup?
  3. You have 300M users. Storing a full trie in memory is expensive.
     How do you shard typeahead across machines?

What to find while reading:
  - What data structure does Cleo use for prefix matching?
  - How does it handle out-of-order query terms ("john smi" matching "John Smith")?
  - How does LinkedIn rank results — why is "John Smith CEO" ranked higher
    than "John Smith intern"?
  - How is the typeahead index updated when a new user joins?

Typeahead design patterns for interviews:
  Simple approach:  trie in memory, single server, prefix lookup O(prefix length)
  At scale:         shard trie by first letter or prefix, replicate for read throughput
  Ranking:          score = frequency × recency × personalization weight
  Update latency:   accept eventual consistency — index updates asynchronous

---

**Uber — Evolving Distributed Tracing at Uber Engineering**
URL: https://www.uber.com/blog/distributed-tracing/
Topic: Observability, distributed tracing across microservices

Pre-reading questions:
  1. A request to Uber's API touches 20 microservices. One of them is slow.
     How do you find which one without checking 20 dashboards?
  2. What is a trace and what is a span in distributed tracing terminology?
  3. You want to trace every request. At 1M requests/sec, storing every trace
     is expensive. How do you decide which traces to keep?

What to find while reading:
  - What specific debugging problem prompted Uber to build Jaeger?
  - What is the difference between a trace and a span?
  - Why did Uber switch from pull-based to push-based trace collection?
  - What is head-based vs tail-based sampling and what tradeoff does each make?

Why observability matters in interviews (mention this unprompted):
  Every system design is incomplete without:
    - Metrics: "trip completion rate dropped 20% in NYC" (dashboard alert)
    - Traces: "which service added 500ms to this request?" (debugging)
    - Logs: "what exactly happened for order #42?" (incident investigation)
  Senior signal: propose these before the interviewer asks.

---

### Week 14 — Notification systems

**Notification system architecture (Uber RAMEN — already in Week 4)**
URL: https://eng.uber.com/real-time-push-platform/
Re-read with a different focus: push notification delivery to offline devices.

First read (Week 4) focus: WebSocket fan-out to online clients.
Second read (Week 14) focus: APNs/FCM delivery to offline/mobile devices.

Pre-reading questions for second read:
  1. A user's phone is locked. Your app needs to notify them of a new message.
     How does the notification reach the device?
  2. APNs (Apple) and FCM (Google) are intermediaries. What do they guarantee
     about delivery — at-most-once, at-least-once, or exactly-once?
  3. You send 10M push notifications simultaneously (flash sale). How do you
     prevent APNs/FCM from rate-limiting or dropping your notifications?

What to find while reading (second pass):
  - How does Uber route a notification to the correct APNs/FCM endpoint?
  - What happens when a device is offline for 3 days — are notifications queued?
  - How do they handle 70,000 pushes/second without overwhelming APNs/FCM?
  - What is the difference between a data message and a display notification?

Notification system design patterns for interviews:
  Online user:   deliver via WebSocket directly (no APNs/FCM needed)
  Offline user:  store in DB + send via APNs/FCM → device wakes → fetches from DB
  Fan-out:       Kafka → Notification Service → APNs/FCM (parallel, batched)
  Deduplication: APNs/FCM may deliver twice — notification_id deduplicates on device
  Retry:         APNs/FCM has its own retry — don't double-retry from your service

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

## After Reading — Write 3 Sentences

For every post, write in your own words:
  1. The problem they had (one sentence, specific — not "they needed to scale")
  2. What they built and the core mechanism (one sentence)
  3. The tradeoff they accepted (one sentence — what did the solution cost them?)

If you cannot write these three sentences, you have not understood the post.
Store these in a file next to this one: interview/reading-notes.md
