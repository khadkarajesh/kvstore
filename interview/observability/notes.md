# Observability — Senior System Design Interview Notes

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
## 1. WHY OBSERVABILITY MATTERS AT SENIOR LEVEL
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

→ Cannot fix what you cannot see
→ Senior interviewers expect you to raise this WITHOUT being asked
→ Absence of monitoring = interviewer marks you as someone who has never operated production systems
→ Bring it up after designing any service: "In production I would also instrument this with..."

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
## 2. THE THREE PILLARS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

METRICS
  → numerical measurements over time: latency, error rate, QPS, CPU
  → when to use: alerting, dashboards, trends
  → aggregate signal — "p99 latency is 800ms right now"

LOGS
  → timestamped records of discrete events
  → example: "user 42 logged in at 14:32 from IP 1.2.3.4"
  → when to use: debugging specific events, audit trails
  → high volume, stored in log aggregator (e.g. ELK, CloudWatch)

TRACES
  → end-to-end journey of a single request across multiple services
  → each service adds a span (start_time, end_time, service_name, parent_trace_id)
  → when to use: finding which service in a chain is slow
  → produces waterfall diagram: shows latency contribution per service

KEY DISTINCTION
  → Metrics tells you THAT something is wrong
  → Logs tells you WHAT happened
  → Traces tells you WHERE in the call chain it happened

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
## 3. THE RED FRAMEWORK (every service)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

R → Rate:     requests per second arriving at the service
E → Errors:   error rate (%) — 5xx responses, panics, timeouts
D → Duration: latency — report p50, p95, p99 (not average)

Rule: state these three metrics for every service you design in an interview.
      "I would instrument this service with RED metrics on every endpoint."

WHY p99 MATTERS MORE THAN p50
  → p50 = median: 50% of requests are faster than this
  → p99 = 1 in 100 requests takes this long
  → At 10,000 RPS → 100 users/sec hitting the p99 tail
  → Average hides tail latency. p99 reveals user-facing pain.
  → Long-tail latency compounds: if service A calls B, C, D → overall p99 is worse than any individual

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
## 4. THE USE FRAMEWORK (every resource / infrastructure)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

U → Utilization: % of time the resource is busy (CPU at 85%, disk at 70%)
S → Saturation:  how much work is queued / waiting (run queue depth, disk I/O wait)
E → Errors:      error rate of the resource (disk errors, network drops, memory ECC errors)

Apply USE to: CPU, memory, disk, network interfaces, database connection pools

Example
  → CPU utilization 40% (fine), saturation 0 (fine), errors 0 (fine) → healthy
  → CPU utilization 95% (high), saturation 5 (queue building) → investigate

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
## 5. ALERTING STRATEGY
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

ALERT ON SYMPTOMS, NOT CAUSES

  → GOOD: "p99 latency > 500ms"  ← users are affected right now
  → GOOD: "error rate > 1%"      ← users are getting errors right now
  → BAD:  "CPU > 80%"            ← cause — may not affect users at all
  → BAD:  "disk > 70%"           ← cause — you have time, no user impact yet

TWO ALERT LEVELS
  → Page (PagerDuty, wake someone up): immediate human response needed
      example: error rate > 5%, p99 > 2s, service completely down
  → Ticket (Jira, fix next sprint): not urgent but needs attention
      example: disk at 60%, minor p99 regression

ALERT FATIGUE
  → definition: too many low-signal alerts → engineers start ignoring them
  → consequence: real incident buried in noise → missed
  → fix: ruthlessly delete alerts that fire without user impact
  → goal: every page should require immediate action; if not, demote it to ticket

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
## 6. DISTRIBUTED TRACING
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

PROBLEM
  → request crosses 5 services, total latency 800ms — which service caused it?
  → logs in each service are disconnected — no way to correlate them

SOLUTION
  → generate unique trace_id at the entry point (API gateway or first service)
  → propagate trace_id in every downstream RPC call header
  → each service logs a span: {trace_id, span_id, parent_span_id, service, start_time, end_time}
  → tracing system assembles spans into a waterfall diagram

WATERFALL DIAGRAM
  ServiceA: [═══════════════════════════════════] 800ms total
    ServiceB: [═══════]                            200ms
    ServiceC:         [═══════════════════════════] 550ms  ← bottleneck
      ServiceD:               [═════]              200ms

TOOLS CATEGORY (know the category, not internals)
  → Open source: Jaeger, Zipkin
  → Cloud: AWS X-Ray, Google Cloud Trace, Azure Monitor

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
## 7. HEALTH CHECKS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

LIVENESS CHECK
  → question: is the process alive?
  → action on failure: restart the container / pod
  → endpoint: GET /healthz → 200 if process is running

READINESS CHECK
  → question: is the process ready to serve traffic?
  → action on failure: remove from load balancer rotation (stop sending requests)
  → endpoint: GET /readyz → 200 only if dependencies are healthy, cache is warm, etc.

KEY DISTINCTION
  → a process can be ALIVE but NOT READY
  → example: process started but still loading 10GB of data into memory cache
  → liveness = should I restart it? / readiness = should I send it traffic?
  → if you only have liveness: load balancer keeps sending requests to unready pods → errors

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
## 8. WHAT TO SAY IN A SENIOR INTERVIEW
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

After designing any service, say:

  "In production I would instrument this with:
   - RED metrics on every endpoint (rate, error rate, p99 latency)
   - Distributed trace IDs propagated through all RPC calls
   - Liveness and readiness health checks on every pod
   - Alerts on p99 latency and error rate, not on CPU or disk
   - Two alert tiers: page for immediate user impact, ticket for slow-burn issues"

This one paragraph separates senior from mid-level candidates.
Mid-level waits to be asked. Senior brings it up proactively.
