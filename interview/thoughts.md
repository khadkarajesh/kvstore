# Thoughts — Decision Frameworks & Mental Models

A running log of decision frameworks built during study sessions.
Use these when an interviewer asks "how would you choose between X and Y?"

---

## Redis pub/sub vs Kafka

### Decision Questions

**1. Need replay on consumer failure?**
- Yes → Kafka
- No → Redis
- Why: Kafka retains messages on disk. Consumer tracks offset — on restart, resumes from last committed offset. Redis pub/sub has no storage. Missed messages are gone.

**2. Need to buffer between producer and consumer?**
- Yes → Kafka
- No → Redis
- Why: Producer writes 100k events/sec. Consumer processes 10k/sec. Kafka absorbs the difference — consumer catches up at its own pace. Redis has no queue — if consumer is slow, messages are dropped or Redis memory fills up.

**3. Need days/weeks of message retention?**
- Yes → Kafka
- No → Redis
- Why: Kafka is disk-backed — cheap to retain large volumes. Redis is memory-backed — expensive at scale.

**4. Need per-partition ordering?**
- Yes → Kafka
- No → Redis
- Why: Kafka guarantees messages with the same partition key arrive in write order within a partition. Redis pub/sub has no ordering guarantee.

**5. Need exactly-one processing per message across a consumer fleet?**
- Yes → Kafka consumer groups
- No → Redis
- Why: Kafka consumer group = one message goes to exactly one consumer in the group. Redis pub/sub = one message goes to ALL subscribers simultaneously.

**6. Need sub-millisecond fan-out to all connected servers?**
- Yes → Redis pub/sub
- No → either works
- Why: Redis is pure in-memory — hash map lookup + TCP socket write = ~0.1-0.5ms. Kafka adds disk write + replication = 2-10ms.

**7. Need to broadcast to all current subscribers simultaneously?**
- Yes → Redis pub/sub
- No → either works
- Why: Redis pub/sub is a broadcast — every subscriber gets every message. Use for SSE/WebSocket routing where all servers need the event to find the right open connection.

**8. Is this a point-to-point real-time notification (SSE, WebSocket routing)?**
- Yes → Redis pub/sub
- No → evaluate other questions

---

### Decision Tree (quick reference)

```
Need replay on consumer failure?          Yes → Kafka
Need to buffer producer/consumer speed?   Yes → Kafka
Need days of message retention?           Yes → Kafka
Need per-partition ordering?              Yes → Kafka
Need exactly-one processing per message?  Yes → Kafka consumer groups

Need sub-millisecond broadcast?           Yes → Redis pub/sub
Need all servers to receive the message?  Yes → Redis pub/sub
Routing real-time events to SSE/WS?       Yes → Redis pub/sub
No persistence needed?                    Yes → Redis pub/sub
```

---

### Why Redis is faster than Kafka

Redis pub/sub (~0.1–0.5ms):
1. Receives PUBLISH command
2. Hash map lookup → find subscriber list (in memory)
3. Write message to each subscriber's TCP socket
No disk. No replication. No serialization overhead.

Kafka (2–10ms):
1. Receives PRODUCE request
2. Writes message to WAL on disk (leader broker)
3. Replicates to follower brokers for durability
4. Acknowledges producer after replication (if acks=all)
Disk write + network replication = higher latency. Still fast, but not sub-millisecond.

---

### Fan-out behavior (common interview confusion)

Redis pub/sub — broadcast:
  PUBLISH "channel" event → [Server1 ✓] [Server2 ✓] [Server3 ✓]
  All subscribers receive every message simultaneously.
  Use when: all servers need the message (SSE/WS routing).

Kafka consumer group — partitioned:
  PRODUCE event → [Consumer1 ✓] [Consumer2 -] [Consumer3 -]
  Exactly one consumer in the group receives each message.
  Use when: parallel processing without duplication.

Kafka multiple consumer groups — broadcast to groups:
  PRODUCE event → Group A (one consumer gets it)
               → Group B (one consumer gets it)
               → Group C (one consumer gets it)
  Each group gets all messages. Within each group, one consumer processes each.
  Use when: analytics + notifications + audit log all need the same event.

---

### Exactly-once per consumer group — real-world significance

The question: "if I scale to 3 consumers, does each message get processed once or three times?"

Redis pub/sub — every subscriber gets every message:
  PUBLISH "orders" {order_id: 42}
  → Consumer 1 receives it → sends confirmation email
  → Consumer 2 receives it → sends confirmation email
  → Consumer 3 receives it → sends confirmation email
  Result: user receives 3 emails. Wrong.

Kafka consumer group — exactly one consumer per message:
  PRODUCE to "orders" topic
  → Consumer 1 receives it (owns the partition) → sends email
  → Consumer 2 receives nothing
  → Consumer 3 receives nothing
  Result: user receives 1 email. Correct.

Real-world impact:

  | Operation              | Redis pub/sub         | Kafka consumer group   |
  |------------------------|-----------------------|------------------------|
  | Send confirmation email| 3 emails sent (wrong) | 1 email sent (correct) |
  | Charge credit card     | Charged 3x (disaster) | Charged once (correct) |
  | Decrement inventory    | -3 stock (wrong)      | -1 stock (correct)     |
  | Route to SSE server    | All 3 check, 1 acts   | Fine — side effect is  |
  |                        | (correct, harmless)   | naturally exclusive    |

Rule:
  Side effect must happen exactly once (charge, email, DB write)?
    → Kafka consumer group

  Operation is read-only or harmless when duplicated (SSE routing,
  cache lookup — only the server holding the connection acts on it)?
    → Redis pub/sub is fine

Why Redis is still correct for SSE routing despite broadcast:
  All 3 servers receive the message. Servers 2 and 3 look up their
  local connection map — user 42 is not there — they discard silently.
  Only Server 1 has the connection and writes to it. Duplicate delivery
  is harmless because the side effect only fires on the one server
  that holds the connection.

---

### One-liner for interviews

Redis pub/sub:
  "Fire-and-forget broadcast to all current subscribers. Sub-millisecond.
  No persistence, no replay. Right choice for routing real-time events
  to open SSE or WebSocket connections."

Kafka:
  "Durable, ordered, replayable event log. Right choice for event-driven
  pipelines, async processing, and any flow where a consumer must survive
  restarts without losing messages."

---
