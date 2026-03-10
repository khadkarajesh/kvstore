## Availability — Flashcards

Format: Q → A
Cover the answer, recall it, then check.

---

**Q: How much downtime does 99% availability allow per year?**
A: 3.65 days per year (7.20 hours/month, 1.68 hours/week).

---

**Q: How much downtime does 99.9% (three nines) allow per year?**
A: 8.77 hours per year (~43.8 minutes/month, ~10.1 minutes/week). Rough anchor: one long outage.

---

**Q: How much downtime does 99.99% (four nines) allow per year?**
A: 52.6 minutes per year (~4.38 minutes/month). Rough anchor: one incident with fast recovery.

---

**Q: How much downtime does 99.999% (five nines) allow per year?**
A: 5.26 minutes per year (~26.3 seconds/month). Near-zero tolerance for planned downtime.

---

**Q: What is the multiplier between each additional nine of availability?**
A: Each additional 9 reduces allowed downtime by ~10x. Going from 99.9% to 99.99% is not "a little better" — it is 10x harder to achieve.

---

**Q: What is an SLI?**
A: Service Level Indicator. The actual measured metric — a ratio of good events / total events over a time window. It is a measurement, not a target. Example: "99.95% of requests completed in <200ms over the last 30 days."

---

**Q: What is an SLO?**
A: Service Level Objective. The internal target for the SLI, owned by engineering. Stricter than the SLA to leave margin before a contractual breach. Breaching an SLO triggers an internal alert with no external consequences.

---

**Q: What is an SLA?**
A: Service Level Agreement. A contractual commitment to a customer with penalties on breach (credits, refunds). The SLA threshold is lower than the SLO — SLA is the public floor, SLO is the internal ceiling.

---

**Q: What is the relationship between SLI, SLO, and SLA?**
A: SLI is what you measure. SLO is what you target internally. SLA is what you promise publicly. SLA <= SLO < 100%. The SLO must be met for the SLA to be safe.

---

**Q: What is an error budget?**
A: error budget = 100% - SLO. For a 99.9% SLO, the budget is 0.1% = 8.77 hours/year = 43.8 minutes/month. It represents the total allowed downtime/error rate before the SLO is breached, shared between reliability work and feature velocity.

---

**Q: What happens when an error budget is exhausted?**
A: All non-critical feature releases are frozen. Only reliability improvements and bug fixes may be deployed. Feature work resumes when the budget replenishes the next month.

---

**Q: What does a 5x burn rate mean, and what action should it trigger?**
A: The system is consuming the error budget five times faster than the SLO allows. Page on-call immediately. A 2x burn rate warrants investigation; 1x means on track.

---

**Q: How does availability compound across synchronous dependencies?**
A: Multiply each dependency's availability. Two services at 99.9% each: 99.9% x 99.9% = 99.8%. Ten services each at 99.9%: 0.999^10 ≈ 99.0% — your service is down to two nines even though every dependency is three nines.

---

**Q: How does using an async message queue improve availability compared to a synchronous RPC call?**
A: With async, the downstream service being unavailable does not cause the caller to fail — the message is queued and delivered later. With synchronous RPC, downstream failure propagates directly to the caller, compounding unavailability.

---

**Q: What is the availability of two redundant parallel paths, each at 99.9%?**
A: 1 - (1 - 0.999) x (1 - 0.999) = 1 - 0.000001 = 99.9999%. Redundancy dramatically improves availability because both paths must fail simultaneously.

---

**Q: What is p99 latency?**
A: The latency below which 99% of requests complete. The slowest 1 in 100 requests exceeds this value. At 100,000 QPS, a p99 of 500ms means 1,000 users per second are experiencing a slow response.

---

**Q: Why does average latency hide problems that p99 exposes?**
A: Averages are dominated by the fast majority. Example: 990 requests at 50ms and 10 at 5,000ms gives an average of ~99.5ms — looks fine, but 10 users per thousand wait 5 seconds. p99 surfaces these tail experiences.

---

**Q: Name three common causes of tail latency spikes.**
A: (1) GC pauses (JVM/Go stop-the-world, 10-100ms). (2) Hot partition — one shard gets disproportionate load, queue builds. (3) Retry storm — failed requests retry, doubling load and cascading slowness. Also: lock contention, TCP retransmit on network jitter.

---

**Q: What is a hedged request and when is it used?**
A: Send the same read request to two replicas simultaneously and take whichever responds first. This cuts p99 latency dramatically at the cost of doubling read load. Acceptable for critical reads where tail latency is intolerable.

---

**Q: What infrastructure is required to achieve 99.999% availability?**
A: Multi-region active-active deployment with a global load balancer. Database must support multi-region replication (eventual consistency) or global strong consistency (e.g., Spanner, CockroachDB). Chaos engineering, Game Days, and an on-call rotation with <5 minute response SLA are required. An entire region failure must be absorbed by other regions with no downtime.

---
