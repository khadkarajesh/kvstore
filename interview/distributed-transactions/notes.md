Distributed Transactions — FAANG Interview Reference

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. The Problem
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

A single operation spans multiple services or databases.

  Example: "Place order"
    Step 1 → debit account        (PaymentService)
    Step 2 → reserve inventory    (InventoryService)
    Step 3 → send confirmation     (NotificationService)

  Failure scenario: step 2 fails after step 1 succeeds
    → account is debited but no inventory reserved
    → partial failure, money taken, order not placed

  Why standard DB transactions don't work:
    → each service owns its own DB
    → no shared transaction manager across service boundaries
    → 2PC requires all participants to be in the same trust domain

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
2. Two-Phase Commit (2PC) — Brief Review
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Phase 1 — Prepare:
  coordinator → asks all participants: "can you commit?"
  each participant → locks resources, writes redo/undo log, replies YES or NO

Phase 2 — Commit or Abort:
  all YES → coordinator sends COMMIT to all participants
  any NO  → coordinator sends ABORT, all participants rollback

---
Failure Modes

  Coordinator fails after Phase 1, before Phase 2:
    → all participants are LOCKED indefinitely
    → no participant knows whether to commit or rollback
    → blocking protocol — cannot make progress without coordinator

  Participant fails mid-commit:
    → recovery log used to replay commit on restart

---
When to Use 2PC

  USE:
    → single database (internal XA transactions)
    → same team, same org, all participants under your control
    → infrequent transactions, latency not critical

  DO NOT USE:
    → across microservices from different teams
    → external APIs (Stripe, Twilio — you cannot ask them to "prepare")
    → cross-org boundaries
    → high-throughput paths (locks held across network round trips)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
3. Saga Pattern
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Definition:
  Break a distributed transaction into a sequence of local transactions.
  Each step publishes an event on success.
  On failure → run compensating transactions to undo previous steps.

  Key difference from 2PC: no locking across services, no coordinator holding resources

---
Example: Place Order Saga

  Happy path:
    1. OrderService    → create order (PENDING)        → publish OrderCreated
    2. PaymentService  → debit account                 → publish PaymentCompleted
    3. InventoryService→ reserve item                  → publish InventoryReserved
    4. OrderService    → mark order CONFIRMED

  Failure at step 3 (InventoryFailed):
    3a. InventoryService → publish InventoryFailed
    3b. PaymentService   → refund account              (compensating transaction)
    3c. OrderService     → mark order FAILED

---
Compensating Transactions

  Not the same as a rollback.
  A rollback reverts a DB transaction that has not committed yet.
  A compensating transaction is a new forward transaction that semantically undoes a previous committed step.

  Example: PaymentService cannot "undo" a debit — it must issue a refund.
  Every saga step must have a defined compensating transaction designed upfront.

  Compensating transactions must also be idempotent.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
4. Choreography vs Orchestration
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Choreography — event-driven, no central coordinator
  Each service listens for events and knows what to do next.

  OrderCreated → PaymentService listens → debits → publishes PaymentCompleted
  PaymentCompleted → InventoryService listens → reserves → publishes InventoryReserved

  Pros:
    → no single point of failure
    → services are fully decoupled
    → adding a step = add a new service that listens to an event

  Cons:
    → overall saga state is invisible — spread across event log
    → hard to answer "what state is order 42 in?"
    → complex compensation flows are hard to reason about
    → cyclic event dependencies are possible and dangerous

---
Orchestration — central orchestrator directs each step

  SagaOrchestrator:
    → calls PaymentService.Debit()
    → on success, calls InventoryService.Reserve()
    → on failure at any step, calls compensations in reverse order
    → stores saga state in its own DB

  Pros:
    → single place to see full saga state
    → explicit, readable flow
    → easier to debug and test
    → compensation logic is centralized

  Cons:
    → orchestrator is a bottleneck
    → orchestrator failure must be handled (make it durable, not in-memory)
    → couples orchestrator to each participant's API

---
When to Use

  Choreography → simple linear flows, 2-3 steps, teams that own all services
  Orchestration → complex flows with branching and many compensation paths,
                  flows that span teams, when you need audit log of saga state

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
5. Idempotency Requirement
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Every step in a saga must be idempotent — safe to execute multiple times with same result.

Why:
  Network failures → retries
  Orchestrator crashes and restarts → replays last step
  Message broker delivers at-least-once → duplicate events

Implementation:
  Each step stores:  (saga_id + step_id) → result
  On retry: check if result already stored → return it directly, skip execution

Examples of non-idempotent operations (dangerous without deduplication):
  → charge credit card (duplicate charge)
  → send email (duplicate email)
  → increment counter (double-counted)

  Fix: deduplication table keyed by (saga_id, step_id, idempotency_key)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
6. The Outbox Pattern
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Problem: how to atomically write to DB AND publish an event?

  Wrong approach:
    1. Write to DB  ✓
    2. Publish event to Kafka ✗  (crash here → DB write survived, event lost)
    Result: saga is broken — next step never triggers

  Correct approach — Outbox Pattern:
    1. BEGIN TRANSACTION
       → write business record to DB  (order row)
       → write event to outbox table  (same transaction)
    2. COMMIT
    3. Separate relay process (Debezium, poller, or CDC):
       → reads outbox table
       → publishes event to Kafka
       → marks outbox row as published
    4. On relay crash → restart → re-read unpublished outbox rows → re-publish
       → guarantees at-least-once delivery to broker

  Outbox table schema:
    id, saga_id, event_type, payload, created_at, published_at (null until sent)

---
At-Least-Once vs Exactly-Once

  At-least-once: event may be delivered and processed more than once
    → require idempotent consumers to handle safely

  Exactly-once: each event processed exactly once
    → Kafka transactions + idempotent producer + transactional consumer
    → higher complexity, lower throughput
    → rarely worth it if consumers are idempotent

  Default recommendation: at-least-once delivery + idempotent consumers

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
7. What to Say in a Senior Interview
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  "For this multi-step operation I would use the Saga pattern with orchestration
   since the flow has multiple compensation paths and spans multiple teams.
   Each step is idempotent, deduplication keyed on (saga_id, step_id).
   Events are published via the outbox pattern — same DB transaction as the
   business write — guaranteeing at-least-once delivery to the broker.
   Consumers are idempotent so duplicate delivery is safe."

---
Decision Tree

  All participants same DB?       → standard ACID transaction
  All participants same team/org? → 2PC if low throughput, Saga if high throughput
  External APIs or separate teams → Saga (choreography for simple, orchestration for complex)
  Any step non-idempotent?        → add deduplication table before proceeding
  Event publish must be reliable? → outbox pattern, not fire-and-forget
