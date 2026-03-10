Distributed Transactions — Flashcards

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

**Q: What are the two phases of 2PC and what happens in each?**
A: Phase 1 (Prepare) — coordinator asks all participants to lock resources and reply YES/NO.
   Phase 2 (Commit/Abort) — if all YES → COMMIT sent to all; any NO → ABORT sent, all rollback.

---

**Q: What is the critical failure mode of 2PC?**
A: Coordinator crashes after Phase 1 but before Phase 2. All participants hold locks indefinitely
   and cannot decide to commit or rollback without the coordinator. 2PC is a blocking protocol —
   progress halts until coordinator recovers.

---

**Q: When should you NOT use 2PC?**
A: Across microservices from different teams, external APIs (Stripe, Twilio), cross-org boundaries,
   or any high-throughput path. External services cannot participate in a prepare/commit protocol
   and locks held across network round trips kill performance.

---

**Q: What is the Saga pattern?**
A: A sequence of local transactions, each publishing an event on success. On failure, compensating
   transactions execute in reverse order to undo previous steps. No distributed locks, no 2PC.

---

**Q: What is a compensating transaction and how does it differ from a rollback?**
A: A compensating transaction is a new forward business transaction that semantically reverses a
   previously committed step (e.g., issue a refund for a debit). A rollback reverts an uncommitted
   DB transaction. You cannot rollback a committed step — you can only compensate.

---

**Q: What is choreography in a saga?**
A: Each service reacts to events autonomously — no central coordinator. OrderCreated triggers
   PaymentService; PaymentCompleted triggers InventoryService. Decoupled but saga state is invisible.

---

**Q: What is orchestration in a saga?**
A: A central SagaOrchestrator calls each service step by step and handles failures by invoking
   compensations. Saga state stored in orchestrator's DB. Easier to debug, orchestrator is SPOF.

---

**Q: Choreography vs Orchestration — when to use each?**
A: Choreography → simple linear flows (2-3 steps), single team, no complex branching.
   Orchestration → complex flows with branching, multiple teams, many compensation paths,
   or when you need a queryable audit log of saga state.

---

**Q: Why must every saga step be idempotent?**
A: Network failures cause retries. Orchestrators crash and replay last step on restart. Message
   brokers deliver at-least-once. A non-idempotent step (e.g., charge card) executed twice causes
   incorrect behavior (duplicate charge). Idempotency makes retries safe.

---

**Q: How do you implement idempotency in a saga step?**
A: Store (saga_id + step_id) → result in a deduplication table. On each execution, check the
   table first. If result exists, return it without re-executing. On first execution, store result
   and return it.

---

**Q: What problem does the outbox pattern solve?**
A: The dual-write problem: atomically writing to a DB AND publishing an event.
   Without it: DB write succeeds, then publish fails → event lost, saga broken.

---

**Q: How does the outbox pattern work?**
A: In the same DB transaction: write business record + write event to outbox table.
   Separate relay process (CDC/poller) reads unpublished outbox rows and publishes to broker.
   On relay crash → restart → re-publish unpublished rows → at-least-once delivery guaranteed.

---

**Q: What is at-least-once delivery?**
A: An event may be delivered to a consumer more than once. Requires idempotent consumers
   to handle duplicates safely. Default guarantee from most message brokers (Kafka, SQS).

---

**Q: What is exactly-once delivery and when is it worth the cost?**
A: Each event processed exactly once. Requires Kafka transactions + idempotent producer +
   transactional consumer. Higher complexity and lower throughput. Rarely worth it — prefer
   at-least-once delivery with idempotent consumers.

---

**Q: Why can't you just use a standard DB ACID transaction across microservices?**
A: Each microservice owns its own database. There is no shared transaction manager across
   service boundaries. ACID transactions require all participants to be in the same DB
   (or use 2PC, which requires all to be in the same trust domain and control plane).

---

**Q: An order saga has 4 steps. Step 3 fails. What happens?**
A: Compensating transactions execute in reverse order: step 2 compensation runs, then
   step 1 compensation runs. Step 4 never executed. Final state: all completed steps
   are semantically undone, system in consistent state.

---

**Q: What is the outbox table schema?**
A: id, saga_id, event_type, payload, created_at, published_at (NULL until published).
   Relay queries WHERE published_at IS NULL, publishes, then sets published_at = now().

---

**Q: Senior interview script: how do you handle a multi-step operation spanning services?**
A: "I'd use the Saga pattern with orchestration for complex compensation paths across teams.
   Each step is idempotent, deduplication keyed on (saga_id, step_id). Events published via
   the outbox pattern — same DB transaction as the business write — guaranteeing at-least-once
   delivery. Consumers are idempotent so duplicate delivery is safe."
