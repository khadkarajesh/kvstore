# Observability — Flashcards

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

**Q: What are the three pillars of observability?**
A: Metrics (numerical measurements over time), Logs (timestamped discrete events), Traces (end-to-end request journey across services).

---

**Q: When do you use metrics vs logs vs traces?**
A: Metrics → alerting, dashboards, trends. Logs → debugging specific events, audit trails. Traces → finding which service in a chain is slow.

---

**Q: What does RED stand for and where does it apply?**
A: Rate (requests/sec), Errors (error rate %), Duration (latency). Apply to every service you design.

---

**Q: What does USE stand for and where does it apply?**
A: Utilization (% busy), Saturation (work queued), Errors (resource error rate). Apply to every infrastructure resource: CPU, memory, disk, network.

---

**Q: Should you alert on "CPU > 80%"? Why or why not?**
A: No. Alert on symptoms, not causes. CPU at 80% may not affect users at all. Alert on p99 latency > 500ms or error rate > 1% — those directly indicate user impact.

---

**Q: Give an example of a good alert and a bad alert.**
A: Good: "p99 latency > 500ms" (symptom — users affected). Bad: "CPU > 80%" (cause — no confirmed user impact).

---

**Q: What is alert fatigue and why is it dangerous?**
A: Too many low-signal alerts → engineers start ignoring them → real incident is buried in noise and missed. Fix: delete alerts that fire without requiring immediate action.

---

**Q: What are the two tiers of alerts?**
A: Page (wake someone up — immediate user impact, needs action now) and Ticket (fix next sprint — no immediate user impact, slow-burn issue).

---

**Q: What is a distributed trace ID and why is it needed?**
A: A unique ID generated at request entry and propagated through every downstream RPC call. Without it, logs across services are disconnected — you cannot correlate which service caused high latency on a specific request.

---

**Q: What is a span in distributed tracing?**
A: A single unit of work within a trace. Contains: trace_id, span_id, parent_span_id, service_name, start_time, end_time. Spans assembled together form a waterfall diagram.

---

**Q: What is the difference between a liveness and a readiness health check?**
A: Liveness: is the process alive? Failure → restart it. Readiness: is the process ready to serve traffic? Failure → remove from load balancer rotation. A process can be alive but not ready (e.g. still warming up cache).

---

**Q: What happens if you only implement liveness checks and not readiness checks?**
A: Load balancer keeps sending requests to pods that are alive but not ready (e.g. still initializing) → users get errors during startup or cache warm-up.

---

**Q: What should you proactively say after designing any service in a senior interview?**
A: "In production I would instrument this with RED metrics on every endpoint, distributed trace IDs on all RPC calls, liveness and readiness health checks on every pod, and alerts on p99 latency and error rate — not on CPU."

---

**Q: Why does p99 latency matter more than average (p50) latency?**
A: p50 hides tail latency. At 10,000 RPS, 100 users/sec hit the p99 tail. Tail latency also compounds: a call chain through services A→B→C has a worse overall p99 than any individual service.

---

**Q: What is the difference between metrics and logs?**
A: Metrics are aggregated numerical signals over time (e.g. "error rate is 2% over the last 5 min"). Logs are individual discrete event records (e.g. "request X failed at 14:32 with 500 error").

---

**Q: A request goes through 5 microservices and is slow. How do you find which one?**
A: Distributed tracing. A trace ID is propagated through all services. Each service records a span with its start/end time. The tracing system renders a waterfall diagram showing each service's latency contribution.

---

**Q: What is the RED metric "Duration" and how should you report it?**
A: Duration = latency. Report as percentiles: p50, p95, p99. Never report only average — it hides tail latency. p99 is the most important for user experience.

---

**Q: CPU utilization is 95%, saturation queue depth is 8. What does this mean?**
A: The CPU is busy 95% of the time (high utilization) and there are 8 tasks waiting in queue (non-zero saturation) — the CPU is a bottleneck and requests are being delayed.

---

**Q: What log aggregation tools category should you mention in an interview?**
A: ELK stack (Elasticsearch, Logstash, Kibana), AWS CloudWatch Logs, Splunk. Know the category — centralized log aggregation — not the tool internals.

---

**Q: What distributed tracing tools should you know for interviews?**
A: Know the category and names: Jaeger (open source), Zipkin (open source), AWS X-Ray, Google Cloud Trace. You don't need to know internals — just when to use tracing and what it produces (waterfall diagram).

---
