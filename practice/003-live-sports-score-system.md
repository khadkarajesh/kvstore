# Live Sports Score System

**Scenario:** Design a real-time sports score delivery system. Every goal/point
scored must be delivered to all watching users within 2 seconds.

Requirements:
- 10 million concurrent users watching the same match
- Score update delivered within 2 seconds of the event
- System must tolerate failures without losing score events
- Score ordering must be correct (stale score is wrong)

---

## High-Level Architecture

```
Score Event (goal scored)
        ↓
Score Service
        ↓
Kafka  (durability layer)
        ↓
Score Consumer
        ↓
Redis pub/sub  (delivery layer)
        ↓
Regional Fan-out Servers (10-20 per region)
        ↓
Edge SSE Servers (200+ total)
        ↓
10M Clients (SSE connections)
```

---

## Component Breakdown

### 1. Delivery Mechanism — SSE

SSE is the correct choice. Score updates are server → client only.
Clients never push data back on the same channel. WebSocket adds
bidirectional overhead that is not needed here.

SSE advantages for this use case:
- Unidirectional — server pushes, client only receives
- Auto-reconnect built into browser EventSource API
- Plain HTTP — works through proxies and firewalls
- HTTP/2 multiplexing — multiple SSE streams over one TCP connection

---

### 2. Two Layers — Kafka + Redis (not alternatives, complementary)

This is the most important concept in this design.
Kafka and Redis solve different problems at different layers.

```
Layer 1 — Durability (Kafka):
  "Did this event happen? In what order? Replay if needed."
  - Survives crashes
  - Ordered per partition
  - Replayable on consumer failure
  - Source of truth

Layer 2 — Delivery (Redis pub/sub):
  "Get this event to every connected server right now."
  - Sub-millisecond broadcast to all subscribers
  - Fire-and-forget (no persistence)
  - One PUBLISH → all subscribed servers receive simultaneously
```

Remove Kafka: events lost on Redis failure, no replay on reconnect.
Remove Redis: Kafka consumer delivers to one server (consumer group),
              not broadcast to all 200 servers.
You need both.

Flow:
  Score service → Kafka → Score consumer → Redis PUBLISH "match:10" → all SSE servers

---

### 3. Fan-out — Redis pub/sub

Each SSE server subscribes to the match channel on connect:

  Server 1 → SUBSCRIBE "match:10"
  Server 2 → SUBSCRIBE "match:10"
  Server N → SUBSCRIBE "match:10"

When a goal is scored:
  Score consumer → PUBLISH "match:10" '{"score":"2-1","time":67}'

Redis delivers simultaneously to all subscribed servers.
Each server writes to all open SSE connections it holds.

---

### 4. Stateless SSE Servers

SSE servers hold open HTTP response objects per client connection.
They do NOT hold user-specific session state (no user preferences,
no auth state, no per-user data).

They only hold:
  { match_id → list of open SSE response objects }

This means any server can serve any client. Load balancer can route
new connections to any server freely — no sticky sessions needed.

Load balancer algorithm: Least Connection.
  Routes each new SSE connection to the server with fewest open connections.
  Ensures even distribution of the 10M connections across all servers.

---

## What Breaks at 10 Million Users

Trace one event end-to-end and ask: what does each component do
when 10M users are connected?

```
Goal scored
    ↓
Kafka write — handles millions of writes/sec. Not the bottleneck.
    ↓
Score consumer reads, publishes to Redis
    ↓
Redis PUBLISH "match:10"
    → 200 servers subscribed to this channel
    → Redis pushes to 200 TCP connections simultaneously
    → BOTTLENECK 1: one Redis node, one channel, 200 deliveries per event
    ↓
Each SSE server receives message
    → writes to 50,000 open HTTP response objects simultaneously
    → BOTTLENECK 2: 50k write() syscalls per server per event
    ↓
Client renders update
```

Bottleneck 1 — Redis single-node fan-out:
  One Redis node delivering to 200 servers for every event.
  At 10M users this becomes a Redis CPU and bandwidth problem.

Bottleneck 2 — SSE server write burst:
  Each server must iterate 50k open connections and write to each.
  OS file descriptor limits and CPU time to iterate connections.

---

## Solution — Hierarchical Fan-out

Break the fan-out into stages so no single node does all the work.

```
Score consumer
      ↓
1 Redis node per region
      ↓
10 regional fan-out servers (each subscribes to regional Redis)
      ↓
200 edge SSE servers (each subscribes to a regional fan-out server)
      ↓
10M clients
```

Each hop multiplies receivers. No single node is overwhelmed.

Mental model — same pattern appears everywhere:
  CDN:               1 origin → many edge nodes → millions of clients
  Push notifications: 1 event → APNs/FCM → millions of devices
  Live sports:        1 score → regional Redis → servers → 10M SSE connections

Rule: when a single broadcaster talks to too many receivers directly,
add a fan-out tier between them. Repeat until no single node is overwhelmed.

---

## Redis Channel Sharding vs Hierarchical Fan-out — When to Use Which

These fix different bottlenecks on different resources. They are not
the same thing even though both involve "too much load on Redis."

  CPU bottleneck (channel sharding fixes this):
    Caused by: high event frequency × many subscribers per channel
    Redis work: on every PUBLISH, Redis loops through all subscribers,
                serializes the message, copies into each send buffer.
                This runs in Redis's single-threaded event loop.
                100,000 events/sec × 200 subscribers = event loop saturated.
    Resource:   Redis CPU / event loop time
    Fix:        split "match:10" into "match:10:shard:0..N"
                publisher writes to all shards, each shard covers a
                subset of servers, CPU work spread across shards

  Connection bottleneck (hierarchical fan-out fixes this):
    Caused by: too many persistent TCP connections into Redis
    Redis work: each connected server costs a file descriptor + buffers
                200 SSE servers = 200 file descriptors (fine)
                100,000 servers = 100,000 file descriptors → OS limit hit
                Redis starts rejecting new connections before CPU is a problem
    Resource:   file descriptors + memory, not CPU
    Fix:        Redis talks to 10 regional fan-out servers only
                each fan-out server maintains connections to 20 edge servers
                Redis connection count drops from 200 → 10

Concrete scenarios:

  Low frequency, many subscribers:
    1 event/minute, 10,000 subscribers
    → connection bottleneck (10,000 file descriptors)
    → NOT CPU bottleneck (one PUBLISH per minute)
    Fix: hierarchical fan-out only

  High frequency, few subscribers:
    100,000 events/sec, 10 subscribers (e.g. stock ticker)
    → CPU bottleneck (event loop saturated)
    → NOT connection bottleneck (only 10 file descriptors)
    Fix: channel sharding only

  High frequency, many subscribers:
    100,000 events/sec, 10,000 subscribers
    → both bottlenecks hit
    Fix: channel sharding + hierarchical fan-out

Sports score system falls into "low frequency, many subscribers":
  A goal every few minutes → CPU is not the problem.
  200 SSE servers holding persistent connections → connections matter.
  Hierarchical fan-out is the right fix. Channel sharding is not needed.

---

## Hotspot Analysis

Where hotspots live depends on write vs read volume:

  Few writers, many readers → hotspot on the fan-out (read) side
    Sports score:     1 score service writes → 10M viewers receive
                      Hotspot: Redis channel delivering to 200 servers
    Celebrity tweet:  1 celebrity writes     → 50M followers receive
                      Hotspot: fan-out service, not Kafka write volume

  Many writers, few readers → hotspot on the write side
    IoT sensors:      millions of devices writing to Kafka
                      partitioned by device_type → all thermostats
                      land on one partition → hot partition
                      Fix: device_type + device_id as partition key
                           spreads writes by identity across partitions

---
Hotspot fix rule:
  Write hotspot → add identity cardinality to the partition key
                  (device_type → device_type + device_id)
                  NOT time-based suffixes — time does not distribute
                  identity load, it only creates time-based buckets

  Read/fan-out hotspot → shard the Redis channel
                  "match:10" → "match:10:shard:0", "match:10:shard:1"...
                  each shard covers a subset of SSE servers
                  score consumer publishes to all shards
                  no single Redis channel overwhelmed

---
Sports score specific — why Kafka partition hotspot is NOT the problem:
  The score service has 1-2 writers publishing score events.
  Write throughput is negligible — a match produces maybe 30 events/hr.
  Partitioning by match_id is correct and sufficient.
  The hotspot is entirely on the Redis fan-out side, not Kafka writes.

---

## Ordering

Redis pub/sub delivers messages in the order they are published.
But the ordering guarantee must start at the producer.

Problem: if the score service publishes two events out of order,
Redis faithfully delivers them out of order.

Fix: Kafka enforces ordering per partition. Partition by match_id
so all events for a match go to the same partition, in order.
Score consumer always reads in order → publishes to Redis in order
→ clients receive in order.

Client-side: include a sequence number in each SSE event.
Client detects gaps (seq=1, seq=2, seq=4 → seq=3 missing) and
requests a replay from the server.

---

## Redis Failure Handling

Redis failure scenario:
  - Existing SSE connections stay open — clients remain connected
  - New score events cannot be fanned out — events pause, not lost
  - Kafka retains all events at its consumer offset
  - When Redis recovers: Score consumer replays from Kafka offset
  - Client SSE EventSource auto-reconnects, requests missed events
    via sequence number

High availability:
  Redis Sentinel: automatic failover in seconds
  Redis Cluster: sharded across nodes, no single-node bandwidth limit

---

## Latency Budget (2-second SLA)

```
Goal scored → Score service publishes to Kafka    ~10ms
Kafka consumer reads                               ~10ms
Consumer publishes to Redis                        ~1ms
Redis delivers to SSE servers                      ~1ms
SSE server writes to open connections              ~5ms
Network to client                                  ~50-200ms (varies by region)
Browser renders                                    ~10ms
                                         Total:    ~87-237ms
```

Well within 2-second SLA with regional deployment.
Geo-distribution (users connect to nearest region) keeps network
hop under 100ms for most users.

---

## Decision Tree (questions to ask in interview)

1. Direction: server → client only?
   Yes → SSE. No → WebSocket.

2. Do I need persistence/replay?
   Yes → Kafka as durability layer.

3. Do I need broadcast to all connected servers?
   Yes → Redis pub/sub as delivery layer.

4. Are Kafka and Redis alternatives here?
   No. Kafka = durability. Redis = delivery. Use both.

5. Single broadcaster talking to too many receivers?
   Yes → add hierarchical fan-out tier.

6. Do SSE servers need user-specific state?
   No → stateless servers, any server can serve any client,
   least-connection LB, no sticky sessions.

---

## Interview Signals (Senior Level)

Weak answer:
  "Use Kafka for fan-out because it's more durable than Redis."
  → Conflates durability layer with delivery layer.
  → Kafka consumer groups deliver to one consumer, not broadcast.

Strong answer:
  "Kafka and Redis serve different layers. Kafka is the durability
  layer — ordered, replayable, survives crashes. Redis pub/sub is
  the delivery layer — sub-millisecond broadcast to all SSE servers.
  Remove either one and the system breaks in a different way."

Weak answer:
  "Auto-scale the servers to handle 10M users."

Strong answer:
  "The bottleneck is not server count. Trace the event: one Redis
  node fans out to 200 servers on every goal, and each server
  writes to 50k connections simultaneously. Both are fan-out
  bottlenecks. Fix: hierarchical fan-out — regional Redis nodes
  each talking to ~20 local servers instead of one Redis node
  talking to all 200."

---

## Initial State on Connect

SSE only delivers future events. When a user opens the page, the SSE
stream has not started yet — they need the current score immediately.

Correct pattern:
  Step 1 → REST API:  GET /match/10/score → returns current score state
  Step 2 → SSE:       connection opens, receives all future updates

Without step 1, user sees a blank score until the next goal is scored.
This is the standard pattern for any real-time system — REST for current
state, SSE/WebSocket for delta updates.

Race condition to handle:
  A goal could be scored between step 1 and step 2.
  Fix: SSE stream includes sequence numbers. If REST response returns
  seq=5 and first SSE event received is seq=7, client knows seq=6 was
  missed and requests a replay.

---

## Channel Sharding — Publisher Behavior

When "match:10" is sharded into N channels, the publisher (score consumer)
must write to ALL shards — not one. Every shard carries the same event
because every SSE server needs the event.

```
Goal scored → score consumer:
  PUBLISH "match:10:shard:0"  '{"score":"2-1","seq":6}'
  PUBLISH "match:10:shard:1"  '{"score":"2-1","seq":6}'
  PUBLISH "match:10:shard:2"  '{"score":"2-1","seq":6}'
  PUBLISH "match:10:shard:3"  '{"score":"2-1","seq":6}'
```

Each SSE server subscribes to exactly one shard based on its index:
  shard_id = floor(server_index / servers_per_shard)
  Server 37  (index 37,  50 per shard) → shard 0
  Server 72  (index 72,  50 per shard) → shard 1
  Server 130 (index 130, 50 per shard) → shard 2

Sharding splits the delivery workload across N channels.
It does NOT split the data — every server still gets every event.

---

## Score Consumer Fault Tolerance

The score consumer is a single process reading from Kafka and publishing
to Redis. If it crashes, score events stop flowing.

Fix: run multiple consumer instances in the same Kafka consumer group.

```
Kafka partition for "match:10"
        ↓
Consumer Group "score-fan-out"
  → Consumer A (active, owns the partition)
  → Consumer B (standby)
  → Consumer C (standby)
```

Kafka assigns the partition to one consumer at a time. If Consumer A
crashes, Kafka detects the missed heartbeat (~10s) and reassigns the
partition to Consumer B. Events resume automatically.

Offset commit behavior:
  Consumer reads event from Kafka → publishes to Redis → commits offset.
  If consumer crashes after Redis publish but before offset commit:
    → on restart, replays the same event from Kafka
    → Redis receives duplicate PUBLISH
    → SSE servers receive duplicate event
    → clients deduplicate by sequence number (already have seq=6, discard)
  This is safe because sequence numbers make the delivery idempotent.

---

## L7 Load Balancer Configuration for SSE

SSE connections are long-lived. Standard LB defaults break them.
Two critical settings required:

  proxy_buffering off:
    By default nginx buffers the response until complete before forwarding.
    SSE never completes — client receives nothing until connection closes.
    Must disable buffering so each data: line is forwarded immediately.

  Long read timeout:
    Default LB timeout is ~60s. SSE connections live for hours.
    LB kills the connection after 60s → client reconnects → killed again
    → infinite reconnect loop.
    Set timeout to match expected connection lifetime (hours or infinite)
    for SSE endpoints only. Keep short timeout for regular API endpoints.

  nginx config:
    location /match/events {
        proxy_pass         http://sse_backend;
        proxy_buffering    off;
        proxy_read_timeout 24h;
        proxy_cache        off;
    }

L7 is required (not L4) because these settings are per-endpoint —
L4 operates at TCP level and cannot inspect URLs or content types.

---

## Multiple Concurrent Matches

World Cup scenario: 8 matches simultaneously, each with millions of viewers.

Each match is an independent Redis channel:
  "match:10", "match:11", "match:12" ... "match:17"

Each SSE server subscribes only to the channel for the match
the user connected to watch — not all active match channels.

```
User watching match 10 connects to Server 42:
  Server 42 subscribes to "match:10" (if not already subscribed)

User watching match 11 connects to Server 42:
  Server 42 subscribes to "match:11"

Server 42 internal state:
  {
    "match:10" → [response_A, response_B, response_C, ...]
    "match:11" → [response_D, response_E, ...]
  }
```

Server subscribes to a channel on first user connection for that match
and unsubscribes when the last user watching that match disconnects.

Scale implication:
  8 simultaneous popular matches × 10M viewers each = 80M connections
  Each SSE server now holds connections for multiple matches
  and subscribes to multiple Redis channels simultaneously.
  Hierarchical fan-out becomes more important — regional servers
  handle only the matches popular in their region.

---

## Knowledge Gaps to Watch (filled during study session)

Gap 1 — Kafka vs Redis as alternatives:
  They are complementary layers, not alternatives.
  Always ask: which layer does this component serve?
  Durability layer → Kafka. Delivery layer → Redis pub/sub.

Gap 2 — Shallow bottleneck analysis:
  "Add more servers" is not enough. Trace one event end-to-end.
  Name the specific component that breaks and why.
  Then propose the fix at that specific component.

Gap 3 — Initial state on connect:
  SSE only delivers future events. Always pair with a REST call
  to fetch current state before opening the SSE stream.
  Handle the race condition with sequence numbers.

Gap 4 — Channel sharding publisher behavior:
  Publisher writes to ALL shards, not one.
  Sharding splits delivery workload, not the data itself.

Gap 5 — Consumer fault tolerance:
  Multiple consumers in same consumer group.
  Kafka reassigns partition on crash (~10s).
  Sequence numbers make duplicate delivery safe.

Gap 6 — L7 LB configuration:
  proxy_buffering off — without this client receives nothing.
  Long read timeout — without this connections die every 60s.
