Availability — System Design Notes

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Availability Numbers
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  99%     = 3.65 days   downtime/year  |  7.20 hours/month  |  1.68 hours/week
  99.9%   = 8.77 hours  downtime/year  |  43.8 minutes/month|  10.1 minutes/week
  99.99%  = 52.6 minutes downtime/year |  4.38 minutes/month|  1.01 minutes/week
  99.999% = 5.26 minutes downtime/year |  26.3 seconds/month|  6.05 seconds/week

  Rule: each additional 9 reduces allowed downtime by ~10x.
  Rule: 99.9% → 99.99% is NOT "a little better" — it is 10x harder to achieve.

  Practical anchor points:
    99%     → unacceptable for any user-facing service
    99.9%   → baseline SLA for most internal tools and B2B services
    99.99%  → typical target for consumer internet (Twitter, Stripe)
    99.999% → required for: telephony, payment rails, air traffic control

---
How to memorize the numbers

  99.9%   → "three nines" → about 9 hours/year → one long outage
  99.99%  → "four nines"  → about 1 hour/year  → one incident with fast recovery
  99.999% → "five nines"  → about 5 minutes/year → near-zero planned tolerance

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SLI / SLO / SLA
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  SLI — Service Level Indicator
    The actual measured metric. A ratio: good events / total events over a time window.
    Examples:
      "99.95% of requests completed in <200ms over the last 30 days"
      "0.02% error rate over the last 7 days"
      "99.8% of writes durably persisted within 1s"
    What it is NOT: a target. It is a measurement. It is the raw number.

  SLO — Service Level Objective
    The internal target for the SLI. Owned by engineering.
    Examples:
      "We aim for 99.9% of requests to succeed each month"
      "p99 latency ≤ 300ms"
    Key property: SLO is stricter than SLA to leave margin before breach.
    SLO breach → internal alert + incident. No external consequences.

  SLA — Service Level Agreement
    Contractual commitment to a customer. Has penalties on breach (credits, refunds).
    Examples:
      "We guarantee 99.9% monthly uptime. If we miss: 10% credit for the month."
      AWS EC2 SLA: 99.99% monthly uptime commitment.
    Key property: SLA threshold is lower than SLO (SLA is the public floor, SLO is the internal ceiling).

  Relationship:
    SLI is what you measure.
    SLO is what you target internally.
    SLA is what you promise publicly.
    SLI < SLO < 100%, SLA ≤ SLO.

  Hierarchy example:
    SLI: measured 99.94% this month
    SLO: internal target 99.9% → SLI passes
    SLA: promised 99.5% → SLI passes by wide margin → no credits owed

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Availability Compounds Across Dependencies
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Rule: availability of a system = product of availability of its synchronous dependencies.

  Two services A → B:
    A = 99.9%, B = 99.9%
    Combined = 99.9% × 99.9% = 99.8%   ← worse than either component alone

  Three services A → B → C:
    99.9% × 99.9% × 99.9% = 99.7%

  Practical implication:
    A microservice calling 10 downstream services synchronously (each at 99.9%):
    0.999^10 ≈ 99.0% — your service is now two nines even though every dependency is three nines.

  Key insight: minimize synchronous dependencies.
    Prefer: async (message queue) over sync (direct RPC call) wherever delivery latency is acceptable
    Async decouples availability: if downstream is down, caller can still succeed (message queued)

  Parallel dependencies (any one succeeds):
    Two redundant paths A and B, system fails only if BOTH fail:
    Combined = 1 - (1 - 0.999) × (1 - 0.999) = 1 - 0.000001 = 99.9999%
    → Redundancy dramatically improves availability.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Designing for a Specific Availability Target
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  99% (2 nines) — Baseline
    Single region is acceptable.
    Single replica database with daily backups.
    Manual restart on failure.
    Acceptable for: personal projects, internal dashboards, dev environments.
    Failure mode: one disk failure or one bad deploy → hours of downtime.

  99.9% (3 nines) — Standard Production
    Redundant application servers behind a load balancer.
    Primary + replica DB with automated failover.
    Health checks with auto-restart (Kubernetes, ECS).
    Rolling deploys (no full restart required).
    Acceptable for: most B2B SaaS, internal tools, read-heavy consumer products.
    Failure mode: primary DB failover takes ~30s → brief outage.

  99.99% (4 nines) — High Availability
    Multi-AZ (availability zone) deployment: app + DB replicated across 2-3 AZs in same region.
    Load balancer in each AZ; AZ failure → traffic shifts automatically.
    No single points of failure in: compute, storage, networking, load balancing.
    Automated failover tested regularly (chaos engineering basics).
    RTO (Recovery Time Objective) < 1 minute.
    Acceptable for: payment APIs, auth services, real-time consumer products.
    Failure mode: AZ goes down → failover completes in <1 min.

  99.999% (5 nines) — Mission Critical
    Multi-region active-active: traffic served from 2+ regions simultaneously.
    Global load balancer (AWS Route53 latency routing, Cloudflare) routes users to nearest region.
    DB: either multi-region replication (eventual consistency) or global strong consistency (Spanner, CockroachDB).
    Chaos engineering: Game Days, regular fault injection, automated resilience testing.
    On-call rotation with <5 minute response SLA.
    Acceptable for: financial infrastructure, telephone networks, healthcare systems.
    Failure mode: entire region fails → other regions absorb traffic, no downtime.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Error Budget
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Definition:
    error budget = 100% - SLO
    SLO = 99.9% → error budget = 0.1% = 8.77 hours/year = 43.8 minutes/month

  What it represents:
    The total allowed downtime / error rate before the SLO is breached.
    Shared resource between engineering reliability work and feature velocity.

  How it is consumed:
    Every incident, every slow deploy, every failed health check consumes budget.
    Track budget burn rate: if burning 10 hours of budget in the first week of the month →
      will breach by week 3 → must reduce risk immediately.

  Policy when budget is exhausted:
    Freeze all non-critical feature releases.
    Only reliability improvements and bug fixes may be deployed.
    Resume feature work when budget replenishes next month.

  Why it matters in interviews:
    Error budget is the mechanism that prevents reliability vs. velocity arguments.
    → Engineering says: "we have budget remaining, safe to deploy"
    → Ops says: "budget is gone, freeze deployments"
    → Neutral, data-driven decision. Neither side can ignore the budget.

  Burn rate alerts:
    1x burn rate: consuming budget at exactly the SLO rate → on track
    2x burn rate: will exhaust budget in half the window → investigate
    5x burn rate: page on-call immediately

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Latency Percentiles — Why p99 Matters More Than p50
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Definitions:
    p50 (median): 50% of requests complete faster than this value. Half your requests.
    p95:          95% of requests complete faster. The 95th slowest in 100 requests.
    p99:          99% of requests complete faster. The slowest 1 in 100 requests.
    p999:         99.9% of requests complete faster. The slowest 1 in 1,000 requests.

  Why p99 matters more than p50:

    At 100,000 QPS:
      p50 = 20ms → 50,000 requests/sec see ≤ 20ms  (good)
      p99 = 500ms → 1,000 requests/sec see ≥ 500ms  (painful for users)
      p99 at 100K QPS = 1,000 users/second experiencing a slow response

    At 1,000,000 QPS:
      p99 = 10,000 users/second seeing slow responses

    Averages hide the distribution:
      avg = (50ms × 990 + 5000ms × 10) / 1000 = 99.5ms — "average looks fine"
      but 10 users/1000 wait 5 seconds → complaint rate, churn, SLA breach

  Tail latency causes:
    GC pause (JVM, Go stop-the-world): spikes of 10-100ms at unpredictable intervals
    Hot partition: one shard/node gets disproportionate load → queue builds → latency spikes
    Retry storm: failed requests retry → double the load → cascading slowness
    Lock contention: one slow DB query holds a lock → other queries queue → latency spike
    Network jitter: packet loss on cross-AZ links → TCP retransmit adds ~200ms

  Mitigation strategies:
    Hedged requests: send same request to 2 replicas; take whichever responds first → cuts p99 dramatically
      Cost: doubles read load. Acceptable for critical reads.
    Timeout + retry: short timeout on first attempt → fast retry → p99 bounded by retry timeout
    Request cancellation: if client disconnects, cancel in-flight work to free resources
    Tail latency SLO: set an explicit p99 SLO and alert on it, not just on error rate

  Interview signal:
    Mention p99, not average. Say "tail latency."
    Propose hedged requests if asked to reduce p99 of a read path.
    Identify GC and hot partitions when asked about sources of latency spikes.

