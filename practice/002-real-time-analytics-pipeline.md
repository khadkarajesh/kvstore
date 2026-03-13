# Real-Time Analytics Pipeline (Ride-Sharing)

**Scenario:** Design the analytics backend for a ride-sharing app. Every trip event (started, ended, fare calculated, driver rated) is streamed. Requirements:
- Dashboard: "How many trips completed in the last 5 minutes, by city?"
- Alerting: Notify ops if trip completion rate in any city drops >20% vs the previous hour
- Scale: 500,000 events/second at peak
- Latency SLA: Dashboard data must be no more than 10 seconds stale

---

## High-Level Architecture

```
Uber App
   |
   v
Kafka (topic: trip-events, partition key: trip_id or user_id)
   |
   +---> Flink Job 1: Dashboard Aggregation --> Redis --> Dashboard UI
   |
   +---> Flink Job 2: Alerting Aggregation  --> Flink Keyed State
                                                    |
                                                    +--> Kafka (topic: alerts)
                                                              |
                                                              v
                                                        Notification Service
                                                        (email / Slack / PagerDuty)
```

---

## Component Breakdown

### 1. Kafka — Ingestion Layer

**Partition key:** `trip_id` or `user_id`

Why not `city`? Flink does its own `keyBy(city)` internally — it reshuffles the stream regardless of Kafka partitioning. Kafka partition key governs **distribution and throughput**, not logical grouping. Choosing `city` would cause hotspot partitions (NYC/SF have far more events than rural cities). `trip_id` or `user_id` gives uniform distribution.

**Partition count:** At 500k events/sec, assuming ~10MB/s per partition throughput, you need ~50+ partitions. Size with headroom.

**Retention:** Set retention to cover at least your longest replay window (e.g., 24–48 hours). This is the safety net for Flink recovery.

---

### 2. Flink Job 1 — Dashboard (5-minute counts by city)

**Window type:** Tumbling window, size = 5 minutes

**Pipeline:**
```
source (Kafka)
  -> filter: event_type == "TRIP_COMPLETED"
  -> keyBy(city)
  -> TumblingEventTimeWindow(5 minutes)
  -> aggregate: count()
  -> sink: Redis SET city:{city}:5min_count {count}
```

**Why tumbling, not sliding?** The query is "last 5 minutes" — a fixed, non-overlapping bucket. Sliding windows would overlap and recount events, inflating numbers incorrectly for this use case.

**Late data:** Set watermark with 5-second tolerance. Events arriving after the watermark are discarded (or routed to a side output for monitoring). In a ride-sharing context, a 5-second late tolerance is reasonable given mobile network variability.

**Redis write:** Always write the **absolute count** (`SET`), never increment (`INCR`). This makes the sink idempotent — safe for Flink at-least-once delivery (see Exactly-Once section below).

**Latency SLA:** 5-minute tumbling window + ~1s Flink processing + ~1s Redis write = data is at most ~5 seconds stale within a window. Comfortably within 10-second SLA.

---

### 3. Flink Job 2 — Alerting (20% drop vs previous hour)

**The wrong approach:** 1-hour tumbling window. It only emits at the end of the hour — a crash at minute 45 doesn't alert until minute 60. 15-minute blind spot.

**Correct approach:** Sliding window
- **Size:** 1 hour (compare like-for-like periods)
- **Slide:** 1 minute (emit every minute, alert fires within ~1 minute of a drop)

**Tradeoff:** With size=1hr, slide=1min, each event participates in **60 windows**. At 500k events/sec that's significant state. Mitigations:
- Pre-filter to TRIP_COMPLETED only before windowing (reduces volume ~4-5x)
- Use incremental aggregation (Flink `AggregateFunction`) to avoid buffering raw events in window state

**Comparison logic — keep state in Flink, not in an external DB:**

The naive approach is to store hourly counts in Cassandra and query it from within the Flink operator. Don't do this:
- Remote I/O inside a streaming operator adds latency to every window evaluation
- Cassandra downtime stalls your alert pipeline
- Introduces at-least-once → exactly-once complexity across two systems

**Correct approach:** Use **Flink keyed state** (`ValueState<Long>`) to store the previous window's count per city. Flink state is:
- In-process (zero network hops)
- Fault-tolerant via checkpointing to durable storage (S3/HDFS)
- Automatically partitioned by key (city)

```
// pseudocode
class AlertFunction extends ProcessWindowFunction {
    ValueState<Long> previousHourCount  // Flink managed state

    onWindow(city, currentCount):
        prev = previousHourCount.value() ?: currentCount
        if (prev - currentCount) / prev > 0.20:
            emit Alert(city, currentCount, prev)
        previousHourCount.update(currentCount)
}
```

**Alert delivery:** Flink emits to a Kafka topic (`alerts`). A downstream consumer (notification service) sends to PagerDuty/Slack/email. Kafka decouples the alert pipeline from the notification system — if PagerDuty is slow, it doesn't back-pressure Flink.

---

### 4. Exactly-Once Semantics

**Flink's guarantee depends on the sink.**

Flink achieves exactly-once for supported sinks via **two-phase commit (2PC)**:
1. On checkpoint trigger: Flink pre-commits to the sink (e.g., begins a Kafka transaction)
2. Flink writes checkpoint metadata to durable storage (S3/HDFS)
3. On checkpoint completion: Flink commits the sink transaction

If a task manager crashes and restarts, Flink replays Kafka from the **last checkpoint offset** (not from the beginning). Events between the checkpoint and the crash are reprocessed.

**Kafka sink:** Exactly-once supported natively. Flink uses Kafka producer transactions. Downstream consumers must use `isolation.level=read_committed`.

**Redis sink:** No 2PC support. Flink delivers **at-least-once** to Redis. On recovery, window results are rewritten. This is safe **only if writes are idempotent** — always `SET` the absolute count, never `INCR`.

**End-to-end exactly-once path (Kafka → Flink → Kafka):**
```
Kafka (source) → Flink (stateful) → Kafka (alerts topic)
```
This entire path supports exactly-once via Flink checkpointing + Kafka transactions.

---

### 5. Scalability

| Component | Scaling lever |
|---|---|
| Kafka | Add partitions; partition key = trip_id for even distribution |
| Flink Job 1 | Parallelism = partition count; add task managers horizontally |
| Flink Job 2 | Same; sliding window state is heavier — size task managers accordingly |
| Redis | Read replicas for dashboard; shard by city if needed |

**Hotspot mitigation:** Even with trip_id as Kafka key, a sudden surge in one city is handled at the Flink layer — `keyBy(city)` subtask for NYC will be hot. Mitigation: increase Flink parallelism so each subtask handles fewer keys, or pre-aggregate at ingestion with a two-stage aggregation (partial count per partition → combine by city).

---

### 6. Storage for Historical Queries (optional cold path)

If the requirement extended to "trips completed last 30 days by city":

```
Flink → Parquet files on S3 (via batch sink every 5min)
              |
              v
        Athena / Trino (ad-hoc SQL over S3)
```

Keep Redis for hot dashboard data (TTL = 10 minutes). S3 + Athena for cold analytical queries. This avoids over-engineering Cassandra for a query pattern it handles poorly (aggregations across many time ranges).

---

## Key Tradeoffs Summary

| Decision | Choice | Why | Tradeoff |
|---|---|---|---|
| Kafka partition key | trip_id / user_id | Even distribution | Flink must keyBy internally |
| Dashboard window | Tumbling 5min | Matches query semantics | Data up to 5min stale |
| Alert window | Sliding 1hr/1min | Detect drops within 1min | 60x state vs tumbling |
| Alert state | Flink keyed state | No external dependency, fault-tolerant | State lives in Flink cluster |
| Redis write | Absolute SET | Idempotent, safe for at-least-once | Overwrites on recovery (correct) |
| Alert delivery | Kafka → consumer | Decouples Flink from notifier | Extra hop, slight latency |

---

## Interview Signals (Senior / Tech Lead Level)

A senior candidate is expected to:
- Distinguish Kafka partition key (distribution) from Flink keyBy (logic) — these are independent concerns
- Know that tumbling windows are wrong for continuous alerting — and know *why* (emit only at boundary)
- Default to Flink keyed state for intra-job comparisons, not external DB lookups
- Know Flink's exactly-once is sink-dependent (Kafka yes, Redis no) and the mechanism is 2PC + checkpointing
- Identify the idempotency requirement on Redis writes and articulate *why* (at-least-once + replay)
- Size Kafka partitions against throughput, not just mention "add more partitions"
- Propose a cold path (S3 + Athena) when historical queries are mentioned, rather than trying to serve everything from Redis or Cassandra
