## Dynamo Paper — Interview Questions

These are open-ended questions as they would appear in a real system design interview.
Answer each out loud or in writing before checking your notes.

---

### Business Motivation

1. Why did Amazon build Dynamo as an AP system instead of CP? What specific business problem does that solve?

2. What does "always writable" mean and why is it a hard requirement for a shopping cart?

---

### Partitioning & Consistent Hashing

3. You have 1000 nodes in a distributed system. How do you decide which node stores a given key? What problem does naive hashing have at this scale?

4. Why does consistent hashing with only physical nodes create hotspots? How do virtual nodes fix this?

5. A physical node goes down. Walk me through exactly what happens to its data with vnodes vs without vnodes.

6. You have three machines: one with 32 cores, one with 16 cores, one with 8 cores. How does Dynamo distribute load proportionally across them?

7. How much data moves when a node joins or leaves the cluster? Why is this important at scale?

---

### Replication

8. A client writes a key. Which nodes store it and how is that determined?

9. What is a preference list? Why does it contain more than N nodes?

10. A client sends a write request to a random node that is not the coordinator for that key. What happens?

---

### Data Versioning & Conflict Resolution

11. Two clients update the same shopping cart simultaneously from different devices. How does Dynamo detect this is a conflict?

12. What is a vector clock? Draw an example of two vector clocks that are in conflict.

13. Dynamo returns two conflicting versions of a shopping cart to the client. What does the client do with them?

14. When would you use Last Write Wins instead of semantic merge? What is the risk?

15. Why does Dynamo detect conflicts on read rather than on write? What does that trade off?

---

### Failures

16. A node goes down mid-write. Dynamo still needs to satisfy W=2 acknowledgements. How?

17. What is hinted handoff? What happens if the substitute node also crashes before delivering?

18. Two replicas have diverged and hinted handoff has failed. How does Dynamo detect and repair this without transferring all data?

19. An entire data center goes down. How does Dynamo ensure replicas survive?

20. What is the chain of failure recovery mechanisms in Dynamo, from least severe to most severe?

---

### Membership & Failure Detection

21. A new node joins the cluster. How does it discover the other nodes?

22. How does Dynamo know a node has failed? Is the detection instant?

23. What is the tradeoff of gossip-based failure detection vs Raft's heartbeat-based detection?

---

### N, W, R Tuning

24. What do N, W, R mean? What does W + R > N guarantee?

25. You are designing a read-heavy leaderboard that can tolerate slightly stale data. How would you tune N, W, R?

26. You are designing a payment system where every read must return the latest write. How would you tune N, W, R? What is the cost?

27. N=3, W=2, R=2. One node goes down — can you still read? Write? Two nodes go down?

---

### Production & Lessons

28. What was the actual conflict rate in Dynamo's shopping cart in production? What does that tell you about building for worst-case scenarios?

29. Dynamo's median read latency is 5ms but p99 is 300ms. Why does this gap exist and why does it matter more than the median?

30. An interviewer asks: "How would you design a highly available key-value store that never rejects writes?" Walk through your design using Dynamo's principles.
