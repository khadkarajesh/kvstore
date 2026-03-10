## Distributed Transactions — Interview Questions

These are open-ended questions as they would appear in a real system design interview.
Answer each out loud or in writing before checking your notes.

---

### Two-Phase Commit (2PC)

1. What is Two-Phase Commit? Walk through both phases with a concrete example involving three participant services.

2. A 2PC coordinator crashes after Phase 1 (all participants have voted YES) but before it sends the Phase 2 commit message. What happens to the participants? How long are they blocked?

3. Why is 2PC called a "blocking protocol"? What does that mean operationally during a coordinator failure?

4. An engineer proposes using 2PC to coordinate a payment debit (PaymentService) with a Stripe API call (external vendor). What is wrong with this and why can you never use 2PC across external APIs?

5. When is 2PC actually the right choice? Give a scenario where it is appropriate and one where it is clearly not.

---

### Saga Pattern

6. A "Place Order" operation spans three services: PaymentService, InventoryService, and NotificationService. Payment succeeds, inventory reservation fails. Walk through the Saga compensation flow step by step.

7. What is the fundamental difference between a Saga compensating transaction and a database rollback? Why does this distinction matter for design?

8. What is the difference between Saga choreography and Saga orchestration? Draw the event flow for each using the order placement example.

9. A product team says: "We want to use choreography because it has no single point of failure." What is the hidden cost of choreography that makes it harder to operate as complexity grows?

10. A Saga orchestrator crashes mid-flow, then restarts. How does it know where to resume? What must be true about every step for this to be safe?

---

### Failure Scenarios

11. Every step in a Saga must be idempotent. Why? Name three specific scenarios where a non-idempotent step causes a real problem.

12. You use the Outbox Pattern to publish events. Walk through the exact sequence of operations — what is written to the DB, what the relay process does, and why it prevents the "write to DB then lose the event" failure mode.

13. An at-least-once message broker delivers a "PaymentCompleted" event twice to InventoryService. InventoryService is not idempotent. What happens, and how do you fix it?

14. Your Saga orchestrator is a stateless in-memory process. It manages 10,000 concurrent sagas. The process crashes. What happens to all 10,000 sagas?

15. A Saga step calls Stripe to charge a credit card. The HTTP call times out — you do not know if the charge succeeded. What must your retry logic do, and what does Stripe's idempotency key do for you here?

---

### Pattern Selection

16. Walk through the decision tree: when do you use a standard ACID transaction, when do you use 2PC, and when do you use a Saga?

17. You are building an e-commerce checkout that spans OrderService, PaymentService, InventoryService, and ShippingService. The flow has multiple possible failure branches. Do you choose choreography or orchestration, and why?

18. What is the difference between at-least-once and exactly-once delivery? Which do you default to and why?

19. A senior engineer says "just use a distributed transaction." What are three follow-up questions you would ask before agreeing or disagreeing with that recommendation?

20. The Outbox table grows unboundedly. What operational process do you add, and what monitoring would you put on it?

---

### Real-world

21. Walk through how you would design a "transfer money between two bank accounts" operation where each account lives in a different service with a different database. Which pattern do you use and why?

22. A notification email is sent as part of a Saga. If the Saga is compensated (order cancelled), should you send a "cancellation email" as a compensating transaction? Is this always safe?

23. You are designing a ride-sharing checkout: charge rider, pay driver, update trip status, send receipts. Sketch the Saga steps, identify which steps need compensating transactions, and flag which steps are hardest to make idempotent.

24. An interviewer asks: "What happens if your Saga is halfway done and the company's entire datacenter goes down?" Walk through what guarantees the Outbox Pattern provides and what it does not.

25. Two engineers debate: "We should use Kafka for Saga events" vs. "We should use direct RPC calls between services." What are the availability and correctness trade-offs of each approach?
