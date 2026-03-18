# Practice Problem 001 — User Activity Log System

Date attempted: 2026-03-12
Self-assessment: Mid-to-senior boundary

---

## The Problem

Design a system to store and serve user activity logs.
Every time a user views a product, adds to cart, or makes a purchase, an event is recorded.

Requirements:
- Write 50K events/sec at peak
- Query: "all events for user X in the last 30 days" — must return in under 200ms
- Query: "all users who purchased product Y today" — analytics team, 1-hour freshness acceptable
- Dashboard updates: marketing wants results within 1 hour (later pushed to 5 min, then 30 sec)

---

## Clarifying Questions — What to Ask and Why

Before designing, ask these three questions in this order.

**1. Is reporting real-time or is a delay acceptable?**
This is the highest-impact question. It splits the architecture in two:
- Delay acceptable (>1 min) → batch/micro-batch pipeline (Kafka → Flink → Cassandra)
- Real-time (<30 sec) → streaming pipeline with low-latency analytical store (Kafka → ClickHouse/Druid/Pinot)
Never over-engineer toward real-time by default. Always ask first.

**2. What is the exact query shape for analytics?**
"All users who purchased product Y today" could mean:
- A count per hour → pre-aggregate in Flink, store summary row
- Raw event list → store raw events, query on read
- Ad-hoc SQL → you need a query engine (Presto, BigQuery), not just Cassandra
The query shape determines whether you store raw events, pre-aggregations, or both.

**3. What is the time range — fixed 30 days or variable?**
This changes partition sizing strategy. Variable range → you cannot hard-code month into partition key.

---

## The Architecture (non-real-time, 1-hour freshness)

```
Producers (app servers)
        |
        v
     Kafka
  (topic: user-events)
  Event schema:
    - user_id
    - product_id (nullable)
    - event_type: view | cart | purchase
    - event_timestamp (epoch ms) ← critical: always include this
        |
        |--- Consumer A: Raw event writer → Cassandra user_events table
        |
        v
      Flink
  (event time processing, watermarks for late arrivals)
  Window: 1-hour tumbling window
  Aggregates: purchase count per product per hour
        |
        v
     Cassandra
  (two tables — one per query pattern)
```

Kafka sits in front always. Even when writing directly to Cassandra, Kafka is the buffer.
Never skip Kafka when ingestion volume can spike. It decouples producers from consumers.

---

## Data Model — The Most Important Part

**Rule: In Cassandra, one table per query pattern. This is not optional.**

You have two queries. You need two tables.

### Table 1: user_events — answers "all events for user X in last 30 days"

```
CREATE TABLE user_events (
    user_id     UUID,
    month       TEXT,          -- format: '2026-03'
    event_ts    TIMESTAMP,
    event_type  TEXT,
    product_id  UUID,
    event_id    UUID,
    PRIMARY KEY ((user_id, month), event_ts)
) WITH CLUSTERING ORDER BY (event_ts DESC);
```

Why `(user_id, month)` as partition key and not just `user_id`?
- Cassandra partition size limit is ~100MB before performance degrades
- A user active for 2 years with 50 events/day = ~36,000 rows = could exceed limit
- Month caps partition size. You rotate automatically. This is called partition sizing.
- Without month: one giant partition per user → read performance degrades silently

Why `event_ts` as clustering key?
- Clustering key defines sort order inside the partition
- Sorted by timestamp DESC → most recent events come out first
- This makes "last 30 days" a range scan on a single partition, not a full table scan

### Table 2: product_events — answers "all purchases of product Y today"

```
CREATE TABLE product_events (
    product_id  UUID,
    date        TEXT,          -- format: '2026-03-12'
    hour        INT,           -- 0-23
    event_ts    TIMESTAMP,
    user_id     UUID,
    event_id    UUID,
    PRIMARY KEY ((product_id, date, hour), event_ts)
) WITH CLUSTERING ORDER BY (event_ts DESC);
```

Why `hour` in the partition key?
- During a flash sale, one product can receive thousands of events per minute
- Partitioning by `(product_id, date)` alone creates a hotspot — all writes for that
  product hit one node
- Adding `hour` spreads writes across 24 partitions per day per product
- Tradeoff: a 7-day aggregation now requires reading 168 partitions (7 days × 24 hours)
  and merging results in the application layer. Accept this tradeoff when the requirement
  is hourly reporting — you never need to read all 168 at once.

Flink writes to BOTH tables simultaneously from the same Kafka stream.
user_events gets raw events. product_events gets raw events filtered to purchases only.

---

## Failure Handling — What to Know Cold

### Flink goes down for 20 minutes

What happens:
1. Kafka buffers events (retention default 7 days — events are not lost)
2. Flink restarts, reads its last checkpoint from persistent storage (RocksDB backed by S3)
3. Flink replays events from the last committed Kafka offset
4. Backlog of 20 minutes is processed

The critical concept: **event time vs processing time**

- Processing time: when Flink processes the event. Unreliable after recovery.
- Event time: the timestamp inside the Kafka event — when it actually happened.
- Flink must process using event time, not processing time.
- Otherwise: events from the backlog get assigned to the wrong window (now instead of 20 min ago)
  and your dashboard shows wrong counts.

The mechanism: **watermarks**
- A watermark is Flink's assertion: "all events up to timestamp T have arrived"
- Late events (arriving after the watermark) can still be included in the correct window
  if they arrive within a configured allowed lateness (e.g. 5 minutes)
- During recovery, Flink replays with correct event timestamps → windows are assigned correctly
  → dashboard counts are accurate once recovery completes

### Kafka consumer lag — the silent killer
If Flink falls behind (slow processing, backpressure), Kafka consumer lag grows.
The dashboard starts showing stale data with no visible error.
This must be monitored. See Observability section below.

---

## Scaling the Write Path

At 50K events/sec:
- Kafka: partition by user_id hash. More partitions = more parallelism. Start with 64 partitions.
- Flink: one task per Kafka partition. Scale Flink workers horizontally.
- Cassandra: partition key design above distributes writes. Add nodes to scale.

At 500K events/sec (10x):
- Kafka partitions → increase to 256+
- Flink → more workers, watch for state backend (RocksDB) becoming the bottleneck
- Cassandra → token-aware load balancing, increase replication factor to 3

---

## Architecture Evolution by Freshness Requirement

This is the key senior-level concept. Know the thresholds.

```
Freshness needed    Architecture
--------------      ------------
> 1 hour            Batch (Spark). Simplest. Cheapest.
5 min - 1 hour      Micro-batch (Flink tumbling windows). Kafka + Flink + Cassandra.
30 sec - 5 min      Streaming (Flink, smaller windows). Same stack, more Cassandra writes.
< 30 sec            Real-time OLAP. Kafka + ClickHouse or Apache Druid or Apache Pinot.
                    Do NOT try to do this with Cassandra — wrong tool.
```

When marketing says "30-second updates":
- Flink windows still work but 30-sec windows mean high-frequency small writes to Cassandra
- More importantly: Cassandra is optimized for known-partition reads, not analytical aggregations
- Switch the analytics read path to ClickHouse/Druid: columnar storage, vectorized execution,
  designed for "sum of purchases grouped by product for last N hours" queries
- Kafka stays. Raw events still go to Cassandra for the user-facing query (under 200ms).
  Only the analytics path changes.
- Tradeoff you must state: ClickHouse is eventually consistent on ingestion (~seconds lag).
  Cassandra gives you stronger consistency per partition.

---

## Observability — State This Unprompted (Senior Signal)

This is what was missing in the interview. A senior candidate raises this without being asked.

**What to monitor on this pipeline:**

  Kafka consumer lag (most important)
    Alert threshold: lag > 100K messages
    Why: lag growing means dashboard is silently falling behind. No error is thrown.
    Tool: Kafka consumer group lag metric, alert in Grafana/Datadog

  Flink checkpoint duration
    Alert threshold: checkpoint taking > 30 seconds
    Why: slow checkpoints mean recovery will be slow after failure. Also indicates
    state backend (RocksDB) is too large and needs tuning.

  Cassandra write latency p99
    Alert threshold: p99 > 10ms
    Why: p99 spike before average spikes. If you alert on average, you are already late.
    Watch for: partition hotspots (one product during a flash sale) appearing as p99 spikes
    on specific nodes.

  Cassandra partition size
    Alert threshold: partition > 50MB (warn before 100MB hard limit)
    Why: oversized partitions cause read timeouts silently. Easy to miss until a user
    complains that their activity history is loading slowly.

  End-to-end pipeline latency
    Measure: time from event_timestamp in Kafka to dashboard showing the result
    Alert threshold: > 2× the expected window size
    Why: this is the single metric that captures all pipeline health in one number.

**What to instrument with distributed tracing:**
  Trace a single event from Kafka producer → Flink processing → Cassandra write.
  Attach a trace_id to the Kafka message header. This lets you debug why a specific
  event appeared late in the dashboard without guessing which stage was slow.

---

## Gaps Identified in This Session

These are the specific things that were weak or missing. Review before next practice.

  1. Vocabulary gap: event time / processing time / watermarks
     You understood the concept but could not name it.
     File to review: interview/real-time/notes.md (event time section)
     Drill: explain watermarks out loud in 60 seconds without notes.

  2. Partition sizing rationale
     You knew month belongs in the partition key but could not explain why.
     The why: Cassandra partition size limit ~100MB. Month caps growth.
     Drill: next time someone asks "why X in the partition key?" answer with
     "because without it, partition size grows unbounded and degrades reads."

  3. Direct-to-Cassandra without Kafka
     You said "directly ingest to Cassandra" for real-time path.
     Always keep Kafka in front. It is the buffer and the replay log.
     Without Kafka: if Cassandra has a slow node, writes back up with no buffer.

  4. Observability — never mentioned unprompted
     This is the single biggest gap between mid-level and senior in this session.
     Fix: end every design with "here is what I would monitor in production."
     Three things minimum: consumer lag, p99 write latency, end-to-end pipeline latency.

---

## What Strong Answers Looked Like in This Session

  Clarifying: "Is this real-time or is a delay acceptable?" — identified correctly
  Hotspot: "Flash sale on product_id creates hotspot, add hour to partition key" — strong
  Tradeoff: "7-day aggregation now requires multi-partition reads — acceptable given hourly requirement" — strong
  Failure: "Kafka buffers, Flink replays from checkpoint" — correct

These are patterns to repeat. They are what senior sounds like.

---

## Re-attempt Checklist

Before attempting a similar problem, verify you can answer these without notes:

  [ ] What is the Cassandra primary key for user_events and why each component?
  [ ] What is the Cassandra primary key for product_events and why hour is in the key?
  [ ] What breaks when Flink goes down and how does event time fix it?
  [ ] At what freshness threshold do you switch from Cassandra to ClickHouse/Druid?
  [ ] Name three things you would monitor on this pipeline and why each one.

If you cannot answer all five, do not attempt a new problem yet.
Review this file first.
