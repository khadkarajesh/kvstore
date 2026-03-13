Distributed Transactions — Senior Engineer Interview Reference

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
WHY THIS IS HARD — The Core Tension
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

In a monolith with one database, a transaction is simple:

  BEGIN TRANSACTION
    debit account
    reserve inventory
    create order
  COMMIT  ← either all happen or none happen

In microservices, each service owns its own database. There is no
shared transaction manager. The moment you call Service B from Service A,
you have exited the ACID boundary.

The question is: how do you make three separate systems agree on an outcome
when any of them can fail at any moment?

This is not a theoretical problem. It is the most common source of production
money bugs, order inconsistencies, and duplicate charges at every major company.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. THE DUAL-WRITE PROBLEM (the root of most bugs)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

The simplest form of the problem. You need to do two writes that must
both succeed or both fail.

Naive approach:
  1. Write to database        ← succeeds
  2. Publish event to Kafka   ← crash here

Result: DB has the record. Kafka has no event. The next service in the
chain never triggers. You have silent data loss.

This is called the dual-write problem. It appears everywhere:
  - write to DB + publish to message queue
  - write to primary DB + write to replica
  - write to DB + update search index (Elasticsearch)
  - charge credit card + record charge in DB

The flip side is equally bad:
  1. Publish event to Kafka   ← succeeds
  2. Write to database        ← crash here

Result: downstream service processed an event for a record that
doesn't exist in the source DB. Ghost data.

You cannot solve this with application code alone. You need a pattern.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
2. THE OUTBOX PATTERN (solution to dual-write)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

The key insight: use the database itself as the reliable intermediate.
One ACID transaction commits both the business record AND the event.
A separate relay process handles the network publish.

Step-by-step:

  Step 1 — Single DB transaction:
    BEGIN TRANSACTION
      INSERT INTO orders (id, status) VALUES (42, 'PENDING')
      INSERT INTO outbox (saga_id, event_type, payload, published_at)
             VALUES (42, 'OrderCreated', '{...}', NULL)
    COMMIT

  Step 2 — Relay process (CDC or poller):
    SELECT * FROM outbox WHERE published_at IS NULL
    → publish to Kafka
    → UPDATE outbox SET published_at = NOW() WHERE id = ?

  Step 3 — On relay crash:
    → restart relay
    → re-read WHERE published_at IS NULL
    → re-publish (at-least-once delivery is fine, consumers must be idempotent)

Why this works: the DB transaction is atomic. Either both the order row
AND the outbox row are committed, or neither is. The relay only publishes
from committed outbox rows. There is no window where one exists without
the other.

Outbox table schema:
  id           UUID/BIGINT primary key
  saga_id      matches the business entity (order_id, payment_id)
  event_type   'OrderCreated', 'PaymentCompleted', etc.
  payload      JSON event body
  created_at   timestamp
  published_at NULL until published, then SET to publish time

Operational concern: outbox table grows unboundedly.
Add a background job that DELETEs WHERE published_at < NOW() - INTERVAL 7 DAYS.
Add monitoring: alert if MAX(created_at) WHERE published_at IS NULL
is more than N minutes old (relay is stuck).

─────────────────────────────────────────────────────────────────
Two ways to implement the relay:

  Polling (simple):
    A background thread queries the outbox table every N seconds.
    Simple to implement. Creates periodic DB load. Latency up to N seconds.
    Good enough for most use cases.

  CDC — Change Data Capture (Debezium):
    Reads the database's replication log (MySQL binlog, Postgres WAL) directly.
    Captures every committed row change in real-time, <1 second latency.
    Zero polling overhead on the DB.
    Requires running Debezium (Kafka Connect plugin) as infrastructure.

REAL EXAMPLE — Shopify:
  Shopify's core product is a sharded MySQL monolith. They needed to get
  order/product/inventory changes into Kafka for dozens of downstream services.
  Adding an outbox table to every service would require changing every write path.
  Instead, they deployed Debezium reading the MySQL binary log. The binlog IS
  the outbox — every committed write is captured automatically, in order, with
  no application code changes. This handled Black Friday / Cyber Monday (BFCM)
  volume — peak write load — with sub-second lag to Kafka.
  Blog: https://shopify.engineering/capturing-every-change-shopify-sharded-monolith

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
3. TWO-PHASE COMMIT (2PC)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

The classical solution to distributed transactions. A coordinator asks
all participants whether they CAN commit, then tells them to DO it.

Phase 1 — Prepare:
  coordinator → ALL participants: "can you commit?"
  each participant:
    - locks all resources it will need
    - writes undo/redo log to disk (so it can recover if coordinator dies)
    - replies YES or NO

Phase 2 — Commit or Abort:
  if all YES → coordinator sends COMMIT to all
  if any NO  → coordinator sends ABORT to all, all rollback

The locking in Phase 1 is the problem. Participants hold locks across
a network round trip — waiting for the coordinator to reply in Phase 2.

─────────────────────────────────────────────────────────────────
The Blocking Failure Mode

  Timeline:
    t=0: coordinator sends PREPARE to all participants
    t=1: all participants reply YES and lock resources
    t=2: coordinator writes COMMIT decision to its own log
    t=3: coordinator CRASHES before sending COMMIT to participants

  Result:
    - All participants are holding locks
    - No participant knows whether to commit or rollback
    - They cannot proceed without hearing from the coordinator
    - The system is BLOCKED until the coordinator recovers

  This is why 2PC is called a "blocking protocol." The lock window is
  the entire Phase 1→Phase 2 network round trip, which can be hundreds
  of milliseconds. At high concurrency, this becomes a throughput bottleneck.

─────────────────────────────────────────────────────────────────
When to use 2PC vs not

  USE 2PC:
    - All participants are in the same organization/team
    - Participants are databases/services you control (not external APIs)
    - Transaction rate is low (internal admin operations)
    - Your DB supports XA transactions natively (PostgreSQL, MySQL)

  DO NOT USE 2PC:
    - Across external APIs — you cannot ask Stripe or Twilio to "prepare"
    - Across microservices from different teams (lock contention, blast radius)
    - High-throughput paths — locks kill performance
    - When one participant is a message broker (Kafka has no 2PC protocol)

REAL EXAMPLE — Uber (what happens when 2PC is too painful):
  Uber built a fulfillment platform using the Saga pattern with
  application-level compensating transactions. Over time, the compensation
  logic grew across dozens of services and became the primary source of
  production incidents — partial failures left trip and driver state out
  of sync, requiring manual ops intervention. Their solution was not to
  fix the sagas. They migrated to Google Cloud Spanner, which provides
  true distributed transactions with external consistency (stronger than
  serializable), outsourcing the 2PC implementation to Google's infrastructure.
  Blog: https://eng.uber.com/fulfillment-platform-rearchitecture/

  The lesson: hand-rolled distributed transaction logic eventually collapses
  under operational weight. If you need true atomicity across entities at scale,
  use a NewSQL database (Spanner, CockroachDB) rather than implementing it yourself.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
4. THE SAGA PATTERN
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

The practical industry answer to distributed transactions when 2PC
is not feasible.

Core idea: break a distributed transaction into a sequence of local
transactions, one per service. Each local transaction commits independently.
If a step fails, run compensating transactions in reverse order to undo
the previously committed steps.

Key difference from 2PC: NO distributed locks. Each service commits and
releases its lock before the next step starts. The system is never blocked
waiting for a coordinator.

─────────────────────────────────────────────────────────────────
Concrete Example: Place Order (3-service saga)

  Happy path:
    1. OrderService     BEGIN TX → create order (PENDING) → COMMIT
                        → emit OrderCreated event
    2. PaymentService   receives OrderCreated
                        BEGIN TX → debit $50 → COMMIT
                        → emit PaymentCompleted
    3. InventoryService receives PaymentCompleted
                        BEGIN TX → reserve 1x item → COMMIT
                        → emit InventoryReserved
    4. OrderService     receives InventoryReserved
                        BEGIN TX → update order (CONFIRMED) → COMMIT

  Failure at step 3 (item out of stock):
    3. InventoryService → emit InventoryFailed
    3b. PaymentService  receives InventoryFailed
                        BEGIN TX → refund $50 → COMMIT    ← compensating transaction
                        → emit PaymentRefunded
    3c. OrderService    receives PaymentRefunded
                        BEGIN TX → update order (CANCELLED) → COMMIT

  Final state: money returned, order cancelled, inventory unchanged.
  No locks were held between services at any point.

─────────────────────────────────────────────────────────────────
Compensating Transactions — NOT rollbacks

This is a common misunderstanding. Know it cold.

  Database rollback:
    - Reverts a transaction that has NOT YET committed
    - The DB undoes all writes as if they never happened
    - No trace in the DB after rollback

  Compensating transaction:
    - A NEW, forward transaction that SEMANTICALLY reverses a
      PREVIOUSLY COMMITTED step
    - The original write is still in the DB (audit trail preserved)
    - You cannot "undo" a charge — you must issue a refund
    - The compensating transaction itself must be designed upfront,
      not as an afterthought

  Every saga step must have its compensating transaction defined
  BEFORE you start building. If you cannot define a compensation,
  you cannot put that step in a saga.

  Steps that are hard to compensate:
    - Send email: you cannot unsend. Compensation = send a follow-up
      "sorry, this was a mistake" email. This is why notification steps
      should be last in a saga — do them after everything else commits.
    - Charge external payment: compensation = refund. Refund can fail too
      (card closed, bank offline). Design for compensation failure.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
5. CHOREOGRAPHY vs ORCHESTRATION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

These are two ways to implement a saga. The difference is: who drives
the flow?

─────────────────────────────────────────────────────────────────
Choreography — decentralized, event-driven

  Each service listens for events and knows what to do next.
  No central coordinator.

  Event flow for Place Order:
    OrderService emits OrderCreated
        ↓
    PaymentService listens → debits → emits PaymentCompleted
        ↓
    InventoryService listens → reserves → emits InventoryReserved
        ↓
    OrderService listens → marks CONFIRMED

  Compensation flow (InventoryFailed):
    InventoryService emits InventoryFailed
        ↓
    PaymentService listens for InventoryFailed → refunds → emits PaymentRefunded
        ↓
    OrderService listens for PaymentRefunded → marks CANCELLED

  Pros:
    - No single point of failure
    - Services are fully decoupled (don't know about each other, only events)
    - Adding a new step = add a new service that subscribes to an event
    - Low latency (no extra hop through an orchestrator)

  Cons:
    - Saga state is invisible. It's spread across the event log. You cannot
      easily answer "what state is order 42 in right now?"
    - Compensation flows get complex fast. Where does InventoryFailed go?
      Which services need to listen to it? If you add a 4th service, you
      need to trace every event that could affect it.
    - Cyclic event dependencies are possible and dangerous (Service A emits
      an event that triggers Service B which emits an event that triggers A)
    - Debugging requires replaying the event log across multiple services

─────────────────────────────────────────────────────────────────
Orchestration — centralized, command-driven

  A SagaOrchestrator calls each service step by step via RPC/command.
  The orchestrator owns the saga state and drives the flow.

  Flow for Place Order:
    SagaOrchestrator:
      1. call PaymentService.Debit(orderId, amount)
      2. on success → call InventoryService.Reserve(orderId, itemId)
      3. on success → call OrderService.Confirm(orderId)
      on InventoryService failure:
        → call PaymentService.Refund(orderId, amount)  [compensation]
        → call OrderService.Cancel(orderId)

  Orchestrator stores saga state in its own DB:
    saga_id | saga_type | current_step | status | created_at | updated_at

  Pros:
    - Single place to see and debug full saga state
    - Explicit, readable flow — the code reads like the business process
    - Compensation logic is centralized
    - Easy to add timeouts and retries per step
    - Easy to build an audit trail / admin UI showing saga status

  Cons:
    - Orchestrator is a bottleneck and a SPOF if not made durable
    - Orchestrator must be stateful and crash-recoverable (store state in DB,
      not in memory)
    - Couples orchestrator to each participant's API
    - More network hops than choreography

─────────────────────────────────────────────────────────────────
When to use which

  Choreography:
    - Simple linear flows, 2-3 steps
    - All services owned by same team
    - No complex branching or conditional compensation
    - You want maximum decoupling between services

  Orchestration:
    - Complex flows with branching, timeouts, conditional logic
    - Flow spans multiple teams (orchestrator is the contract)
    - You need a queryable audit log of saga state
    - You need centralized retry and timeout management

  Hybrid (DoorDash's approach):
    Primary flow = choreography (Kafka events, fast path)
    Fallback = orchestration (Cadence kicks in when primary flow stalls)
    This is the pragmatic production answer: use events for the happy path,
    use an orchestrator to drive stuck sagas to completion.
    Blog: https://careersatdoordash.com/blog/building-reliable-workflows-cadence-as-a-fallback-for-event-driven-processing/

─────────────────────────────────────────────────────────────────
REAL EXAMPLE — Uber Cadence / Temporal

  Uber built Cadence (now open-sourced as Temporal.io) specifically because
  hand-rolling orchestration sagas was too painful. Cadence guarantees:
    - Workflow code executes to completion even if servers crash
    - State is durably checkpointed after every activity (step)
    - Retry logic is built-in (exponential backoff per activity)
    - Compensation is just: call compensating activities in a catch block

  Instead of writing a saga state machine + recovery logic manually,
  developers write ordinary Go or Java code. Cadence handles replay on crash.

  Used at Uber for: trip lifecycle, payment workflows, driver onboarding.
  Used at DoorDash as the fallback orchestrator.
  Blog: https://www.uber.com/blog/open-source-orchestration-tool-cadence-overview/

  Key concept: Cadence uses EVENT SOURCING internally. It persists every
  event in the workflow execution history. On worker crash, it replays
  the history to reconstruct state, then continues from where it left off.
  This is why workflow code must be DETERMINISTIC — same history = same replay.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
6. IDEMPOTENCY — the safety net for all of the above
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Every saga step will be retried. This is not optional — it is a
certainty in any distributed system:
  - Networks fail and time out
  - Orchestrators crash mid-step and replay on restart
  - Message brokers guarantee at-least-once delivery, not exactly-once
  - Load balancer health checks restart crashed workers

A step is idempotent if executing it N times has the same effect as
executing it once.

─────────────────────────────────────────────────────────────────
The Stripe HTTP timeout problem (canonical example)

  Your code calls Stripe to charge a credit card. The call times out.
  You do NOT know:
    - Did the network fail before reaching Stripe? (charge did not happen)
    - Did Stripe process the charge and the response was lost in transit?
      (charge happened, you just don't know)

  If you retry without an idempotency key: possible double charge.
  If you don't retry: possible missed payment (revenue loss).

  Stripe's solution: Idempotency-Key header.
    First call:  POST /charges  Idempotency-Key: order-42-charge-1
    Stripe executes the charge, stores (key → result).
    Retry call:  POST /charges  Idempotency-Key: order-42-charge-1
    Stripe returns the stored result. Charge is NOT executed again.

  This makes a non-idempotent operation (charge card) safe to retry.
  Blog: https://stripe.com/blog/idempotency

─────────────────────────────────────────────────────────────────
Implementing idempotency in your own services

  Pattern: deduplication table keyed by (saga_id, step_id)

    Table: step_results
      saga_id     | step_id      | status   | result_payload | created_at
      order-42    | charge-card  | SUCCESS  | {charge_id: x} | 2024-01-01

  On each step execution:
    1. SELECT * FROM step_results WHERE saga_id = ? AND step_id = ?
    2. If row exists → return stored result immediately (skip execution)
    3. If not → execute step → INSERT result into step_results → return result

  The INSERT + execution should be atomic (single DB transaction).

─────────────────────────────────────────────────────────────────
REAL EXAMPLE — Airbnb Orpheus library

  Airbnb's Orpheus library decomposes every payment API call into three
  atomic phases:

    Phase 1 (Pre-RPC):   record intent to DB — "I am about to charge"
                         NO network calls in this phase
    Phase 2 (RPC):       make the actual external call (Stripe, Braintree)
    Phase 3 (Post-RPC):  record result to DB — "charge succeeded with id=X"
                         NO network calls in this phase

  Each phase is committed independently to the local DB.
  If the service crashes between phases, on restart it checks:
    - Pre-RPC committed, RPC not done → execute the RPC
    - RPC done, Post-RPC not committed → record the result
    - All three committed → saga step is complete, idempotency key guards against retry

  Why no network calls in Pre/Post phases? One incident taught them:
  connection pool exhaustion during a DB-plus-network-call transaction
  caused cascading failures across the payment stack. Strict phase
  separation prevents this.

  Blog: https://medium.com/airbnb-engineering/avoiding-double-payments-in-a-distributed-payments-system-2981f6b070bb

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
7. AT-LEAST-ONCE vs EXACTLY-ONCE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

At-least-once:
  An event may be delivered to a consumer more than once.
  Default guarantee from Kafka, SQS, and most message brokers.
  Correct behavior requires idempotent consumers.

At-most-once:
  An event is delivered at most once. May be dropped entirely.
  Used when losing a message is acceptable (metrics, logs).
  Achieved by not retrying on failure.

Exactly-once:
  Each event is processed exactly once, end-to-end.
  Requires: Kafka idempotent producer + Kafka transactions +
  transactional consumer (read-process-write atomically).
  Higher complexity. Lower throughput. Harder to operate.

Default recommendation: at-least-once + idempotent consumers.
This gives you effectively-exactly-once semantics with much lower
operational complexity.

When exactly-once is worth it:
  - Financial ledger updates where duplicate entries cause regulatory issues
  - Kafka-to-Kafka pipelines with Flink (Flink's 2PC supports this natively)
  - Deduplication cost is prohibitively high (rare)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
8. CONSISTENCY MONITORING — the safety net you can't skip
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Even with sagas + idempotency + outbox, eventual consistency means your
system can get into inconsistent states silently. You need to detect them.

The pattern: offline consistency checker (async audit pipeline)

  1. Stream all state transitions for entity X into a data warehouse
  2. Batch job (hourly or daily) traces every X from creation to terminal state
  3. Any entity that reached an impossible state = consistency bug
  4. Flag it → automated repair workflow or on-call ticket

REAL EXAMPLE — Airbnb payments:
  Airbnb streams all payment state transitions to their data warehouse.
  An offline checker traces every payment from INITIATED to SETTLED.
  It found real money leakage — double refunds and lost revenue — that
  would never have been caught by service-level metrics alone.
  Blog: https://medium.com/airbnb-engineering/measuring-transactional-integrity-in-airbnbs-distributed-payment-ecosystem-a670d6926d22

This is a senior-level signal interviewers look for. Most candidates
describe the write path. Very few mention the consistency verification path.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
9. DECISION FRAMEWORK
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Question 1: Are all participants in the same database?
  YES → standard ACID transaction. Done.

Question 2: Are all participants in the same organization, low-throughput?
  YES → 2PC via XA transactions (or use a NewSQL DB like CockroachDB/Spanner)

Question 3: Across external APIs, multiple teams, or high throughput?
  → Saga pattern

Question 4: Simple linear flow, 2-3 steps, one team?
  → Choreography (event-driven)

Question 5: Complex branching, multiple teams, need audit log?
  → Orchestration (Cadence/Temporal or hand-rolled orchestrator)

Question 6: Need to publish events reliably from a DB write?
  → Outbox pattern (application-level table or Debezium CDC)

Question 7: Any step calls an external API?
  → Idempotency key on every external call. Always.

Question 8: How do you know your eventual consistency is actually converging?
  → Offline consistency checker in your data warehouse

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
10. WHAT SENIOR/STAFF ANSWERS SOUND LIKE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Mid-level answer:
  "I'd use the Saga pattern with compensating transactions and publish events
  via the outbox pattern."

Senior answer:
  "I'd use an orchestration-based saga since this flow has multiple
  compensation branches and spans three teams — choreography would make
  the compensation logic invisible and hard to debug. Each step is
  idempotent with a deduplication table keyed on (saga_id, step_id).
  Events are published via the outbox pattern in the same DB transaction
  as the business write, giving at-least-once delivery to Kafka. I'd
  use Temporal/Cadence rather than hand-rolling retry and state recovery —
  we'd get durable execution for free. For external payment calls, I'd
  include an idempotency key on every request to handle the timeout-unknown
  failure mode. Finally, I'd add an async consistency checker in the
  data warehouse to detect any silent consistency bugs we miss at the
  service level."

The difference: senior answers name the failure modes they're protecting
against, justify each choice, and mention operational concerns
(observability, consistency verification). They also show awareness of
when NOT to hand-roll (use Temporal vs build your own).

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
REAL COMPANY EXAMPLES — QUICK REFERENCE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Stripe — Idempotency as the primary transaction primitive
  Every API endpoint accepts an Idempotency-Key. Charge operations are
  decomposed into atomic phases (pre-charge / charge / post-charge).
  Avoids 2PC entirely. Retry-safe by design.
  https://stripe.com/blog/idempotency

Airbnb — Orpheus idempotency library + offline consistency checker
  Pre/RPC/Post-RPC phase separation. Prevents connection pool exhaustion.
  Offline consistency checker in data warehouse catches silent money bugs.
  https://medium.com/airbnb-engineering/avoiding-double-payments-in-a-distributed-payments-system-2981f6b070bb
  https://medium.com/airbnb-engineering/measuring-transactional-integrity-in-airbnbs-distributed-payment-ecosystem-a670d6926d22

Uber — Abandoned sagas, migrated to Google Spanner
  Saga compensations across fulfillment services caused too many production
  incidents. Migrated to Spanner for external consistency. Lesson: at scale,
  application-level saga maintenance cost exceeds the cost of a NewSQL DB.
  https://eng.uber.com/fulfillment-platform-rearchitecture/

Uber — Cadence (now Temporal)
  Open-source durable execution engine for orchestration-based sagas.
  State persisted as event history; deterministic replay on crash.
  https://www.uber.com/blog/open-source-orchestration-tool-cadence-overview/

DoorDash — Hybrid: Kafka choreography + Cadence orchestration fallback
  Happy path = Kafka events (fast). Stuck workflows = Cadence drives to completion.
  Practical production pattern: choreography for speed, orchestration for correctness.
  https://careersatdoordash.com/blog/building-reliable-workflows-cadence-as-a-fallback-for-event-driven-processing/

Shopify — Debezium CDC as infrastructure-level outbox
  Avoided modifying every service write path. MySQL binlog IS the outbox.
  Handles BFCM peak load with sub-second Kafka lag.
  https://shopify.engineering/capturing-every-change-shopify-sharded-monolith

Netflix — WAL service as outbox for non-transactional datastores
  Cassandra and EVCache don't support ACID transactions. Netflix WAL service
  provides 2PC semantics for multi-partition writes to these stores.
  https://netflixtechblog.com/building-a-resilient-data-platform-with-write-ahead-log-at-netflix-127b6712359a
