## Observability — Interview Questions

These are open-ended questions as they would appear in a real system design interview.
Answer each out loud or in writing before checking your notes.

---

### Three Pillars

1. What are the three pillars of observability? For each, describe what it tells you and give one concrete example of when you would use it.

2. A production alert fires: "error rate is elevated." Which pillar tells you that something is wrong, which tells you what happened, and which tells you where in the call chain it originated?

3. You have a service that processes payments. An engineer says "we log every request." Is that sufficient observability? What are the two things logging alone cannot give you?

4. What is the difference between metrics and logs at the data level — specifically in terms of cardinality, storage cost, and query time? When does high-cardinality data belong in logs rather than metrics?

5. An interviewer asks you to "design the observability layer" for a new service. What do you add to your design, in what order, and how do you proactively raise it without being asked?

---

### RED / USE

6. What are the RED metrics? Apply them to a specific service: a REST API that handles search queries. Name the exact metric for each R, E, and D.

7. You are asked to instrument a new order processing service. Which three metrics do you add first, and why do you report latency as p99 rather than as an average?

8. At 100,000 QPS, your p99 latency is 600ms. How many users per second are experiencing a 600ms or slower response? Why does this matter more than the average latency of 25ms?

9. What are the USE metrics and what types of resources do they apply to? Give an example of a concerning USE reading for a database server's connection pool.

10. A database is at 90% CPU utilization and the run queue depth is 8 (saturation). What does this tell you about the state of the system, and what is your next diagnostic step?

---

### Alerting

11. What is the rule "alert on symptoms, not causes"? Give two examples of a cause-based alert and rewrite each as a symptom-based alert.

12. What is alert fatigue, what causes it, and what is the concrete operational consequence when it occurs?

13. You have two alert tiers: page (immediate) and ticket (next sprint). Classify these alerts correctly: "p99 latency > 2s," "disk usage > 60%," "error rate > 5%," "memory usage > 80%."

14. Your service's error rate is 0.5%. That is below the 1% page threshold. But the absolute number is 5,000 errors per second because you have 1M QPS. Should you page? What does this reveal about threshold-based alerts alone?

15. Describe a burn rate alert in the context of SLO-based alerting. Why is a "5x burn rate" alert more actionable than a simple "error rate > X%" alert?

---

### Distributed Tracing

16. A request passes through five services and the total latency is 900ms. Logs in each service are isolated. How does distributed tracing identify which service is the bottleneck?

17. What is a trace ID and a span? Walk through exactly what each service adds to a trace as a request flows through a system of three services.

18. What does a trace waterfall diagram show? Draw a simple example with three services where the middle service is the bottleneck.

19. A request enters ServiceA, which calls ServiceB and ServiceC in parallel, then waits for both before responding. How does the trace waterfall look different from a sequential call chain?

20. You are adding tracing to a new service. What must you propagate in every outbound RPC call, and what happens to the trace if one service in the chain fails to propagate it?

---

### Incident Response

21. An alert fires at 2am: "p99 latency on the order service has exceeded 3 seconds." Walk through your diagnostic process using the three pillars of observability in order.

22. What is the difference between a liveness probe and a readiness probe in Kubernetes? Give a concrete scenario where a pod is alive but not ready, and explain what happens without a readiness probe.

23. You deploy a new version of your service and within 5 minutes error rate climbs from 0.1% to 8%. What observability data do you look at first, and how does a distributed trace help you determine if the regression is in your service or a downstream dependency?

24. After an incident is resolved, you write a post-mortem. What observability gaps does the post-mortem process typically reveal, and what is the most common outcome in terms of new alerts or dashboards?

25. An interviewer says "how do you know your system is healthy?" Give the complete answer a senior engineer would give, covering metrics, alerts, traces, and health checks in one structured response.
