## Replication — Interview Questions

These are open-ended questions as they would appear in a real system design interview.
Answer each out loud or in writing before checking your notes.

---

### Replication Modes

1. What is the difference between replication and sharding? Why is confusing these two a red flag in an interview?

2. Walk through synchronous replication. What does the leader wait for before ACK-ing the client, and what are the two specific costs of this approach?

3. Walk through asynchronous replication. What does the leader do before replication completes, and what are the two specific risks?

4. What is semi-synchronous replication? Which specific guarantee does it provide that pure async does not, without the full cost of fully synchronous replication?

5. You are designing a system where write durability is critical (financial transactions) but write latency also matters. Which replication mode do you choose, and how do you configure it?

6. What are the three purposes of replication? Give a concrete example of when each purpose is the primary driver of the decision to replicate.

---

### Lag & Consistency

7. What is the "read-your-own-writes" anomaly? Walk through the exact sequence of events that produces it, and give two concrete ways to fix it.

8. What is the monotonic reads anomaly? Walk through the scenario where a user sees a newer value and then an older value across two sequential reads, and explain the fix.

9. What is the causal consistency anomaly? Give the specific example of a user seeing a reply to a post before seeing the post itself, and what causes it.

10. Your service routes all reads to followers and all writes to the leader. A user submits a profile update and immediately reloads their profile page. What do they see and why? How does your system fix this without routing all reads to the leader?

11. How do you monitor replication lag? What threshold do you alert on, and what do you do operationally when a follower exceeds that threshold?

---

### Failover

12. A leader fails. Walk through the automatic failover process step by step: detection, election, reconfiguration. What consensus algorithm is used for the election?

13. What is split-brain in the context of single-leader replication? Describe the exact scenario where two nodes simultaneously believe they are the leader.

14. What are fencing tokens? Walk through how they prevent split-brain from causing data corruption, step by step.

15. What is STONITH? In what situations is it used, and what out-of-band capability does it require?

16. A newly elected leader was behind the old leader by 10 seconds of replication (async replication). What data is permanently lost? How do you minimize this risk during normal operation?

17. A failed leader comes back online after a new leader has been elected. What must happen, and what catastrophic problem occurs if this is handled incorrectly?

---

### Multi-leader

18. What are the three main use cases for multi-leader replication? For each, explain why single-leader would be insufficient.

19. Two leaders accept writes for the same key at the same time. Neither write is wrong. What are the three conflict resolution strategies, and what are the trade-offs of each?

20. What is Last Write Wins (LWW)? What is the specific failure mode caused by clock skew, and what data loss does it produce silently?

21. What are CRDTs? Give two concrete examples of data structures that can use CRDTs, and one example that cannot.

22. A user edits a document offline on their phone. The same user edits the same document on their laptop with a different change. Both devices sync when network is restored. What replication model does this describe, and how is the conflict resolved?

---

### Real-world Failures

23. You have a single-leader database in US-East. A whole AWS availability zone goes down, taking the leader with it. Walk through exactly what happens to writes, how long the outage lasts, and what an async replication setup loses vs. a semi-sync setup.

24. A follower replica falls 5 minutes behind the leader. New reads start going to this follower. What user-facing bugs appear, and which anomaly category does each belong to?

25. An interviewer asks: "How do you design a database layer for a system that needs both high write throughput and strong read consistency?" Walk through your answer, covering single-leader, multi-leader, and leaderless options, and commit to a recommendation with justification.
