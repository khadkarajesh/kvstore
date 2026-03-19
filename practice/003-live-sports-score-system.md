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

## Knowledge Gaps to Watch (filled during study session)

Gap 1 — Kafka vs Redis as alternatives:
  They are complementary layers, not alternatives.
  Always ask: which layer does this component serve?
  Durability layer → Kafka. Delivery layer → Redis pub/sub.

Gap 2 — Shallow bottleneck analysis:
  "Add more servers" is not enough. Trace one event end-to-end.
  Name the specific component that breaks and why.
  Then propose the fix at that specific component.
