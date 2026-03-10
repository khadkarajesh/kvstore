## Availability — Interview Questions

These are open-ended questions as they would appear in a real system design interview.
Answer each out loud or in writing before checking your notes.

---

### SLI / SLO / SLA

1. What is the difference between an SLI, an SLO, and an SLA? Give a concrete example of each for a payment API.

2. Why is the SLO set stricter than the SLA? What would happen if they were equal?

3. An engineer says "our uptime was 99.94% this month." Is that an SLI, SLO, or SLA? What additional information do you need to determine whether the service is healthy?

4. You are designing a new service. An interviewer asks: "What SLO would you set?" Walk through how you would arrive at a specific number rather than just saying "as high as possible."

5. What are good SLIs for a search API? List at least three, and explain why "CPU utilization" is not a good SLI.

---

### Error Budgets

6. Your SLO is 99.9% monthly uptime. How many minutes of downtime is your error budget for the month? What does it mean operationally when that budget reaches zero?

7. An on-call engineer notices the error budget will be exhausted in three days if the current burn rate continues. What actions should the team take?

8. A product manager wants to ship a large feature this week. Your error budget for the month is already 80% consumed with one week remaining. How do you use the error budget framework to make the decision?

9. What is a burn rate alert? Why is a 5x burn rate alert more urgent than a 1x burn rate alert?

10. A team argues: "We haven't had an incident in six months, let's lower the SLO to ship faster." What is wrong with this argument, and what would you say?

---

### Failure Modes

11. You have three services in a chain: A calls B, B calls C. Each has 99.9% availability. What is the combined availability of a request that must pass through all three? What does this imply about synchronous dependency chains?

12. 10 microservices each have 99.9% availability. They all call each other synchronously in sequence. What is the effective availability of the end-to-end request?

13. What is the difference between making two services redundant (parallel) versus sequential? Give the calculation for each with 99.9% components.

14. What are the common causes of tail latency (p99 spikes)? Name at least four, and explain what hedged requests are and when to use them.

15. A user-facing service is technically returning 200 OK but 40% of responses contain incorrect data. Is this an availability issue? How would you detect it?

---

### Redundancy & Recovery

16. Walk through what engineering changes are required to go from 99.9% to 99.99% availability. What specifically must change in the infrastructure?

17. What is the difference between RTO (Recovery Time Objective) and RPO (Recovery Point Objective)? For a payment system, which one is more critical and why?

18. What is chaos engineering and why does it matter for 99.99% availability targets? What would a "Game Day" exercise look like for a payment API?

19. A primary database fails. Your system has async replication with a follower that is 30 seconds behind. Walk through the failover — what data do you lose, and how do you detect and quantify the loss?

20. Your system needs 99.999% availability. What infrastructure changes are required beyond what you would do for 99.99%? What does "multi-region active-active" mean and what consistency trade-off does it require?

---

### Trade-offs

21. A customer reports that your service was slow for 2 minutes but never returned errors. The SLO only measures error rate. Did you breach the SLO? What does this expose about error rate as a sole SLI?

22. p50 latency is 20ms and p99 is 800ms. An engineer says "our average looks fine at 25ms, no problem." What is wrong with this reasoning, and what does p99 at 100,000 QPS mean in terms of users experiencing slow responses per second?

23. You are designing a system where the product manager insists on "five nines" (99.999%). What question do you ask before committing to that SLO, and what is the realistic annual cost of that decision?

24. Two services have the same SLA. Service A uses sync calls to 8 downstream services. Service B uses async messaging to the same services. Which has better effective availability and why?

25. An interviewer asks: "How do you design for availability?" Give a structured answer covering at least four concrete techniques, not just "add more servers."
