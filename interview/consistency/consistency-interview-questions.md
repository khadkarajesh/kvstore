## Consistency Models — Interview Questions

These are open-ended questions as they would appear in a real system design interview.
Answer each out loud or in writing before checking your notes.

---

### Core Concepts

1. What is the fundamental question that separates consistency models from each other?

2. A junior engineer says "just use strong consistency everywhere to be safe." What's wrong with that argument?

3. An interviewer asks you to pick a consistency model for a new feature. What questions do you ask before answering?

---

### Eventual Consistency

4. You are designing a social media platform's like counter. A post has 10,000 likes. Two users like it simultaneously from different data centers. One user sees 10,001, another sees 10,001, but the actual count should be 10,002. What happened and is this acceptable?

5. DNS is an example of eventual consistency. Walk me through why — what gets replicated, how long does it take, and what happens if you read before propagation completes?

6. What mechanisms does Dynamo use to ensure replicas eventually converge?

7. A product catalog shows a price of $49 on one replica and $59 on another after a price update. Is this a bug or expected behavior? How long will it last?

---

### Session Consistency (Read Your Own Writes)

8. A user updates their profile picture. They refresh the page and still see the old picture. What consistency model failed here and why?

9. Walk me through three different ways to implement read your own writes. What are the tradeoffs of each?

10. You implement sticky sessions by hashing user ID to a node. The node goes down. What happens to the consistency guarantee?

11. What is a version token and how does it solve the sticky session node failure problem?

12. Twitter shows you your own tweet immediately after posting but your friend on a different device doesn't see it for 2 seconds. Which consistency model is Twitter using and is this correct behavior?

---

### Causal Consistency

13. A user posts "Anyone want to grab lunch?" on a social feed. Another user replies "I do!" A third user sees the reply before the original post. What consistency model was violated?

14. How do vector clocks enforce causal consistency? Walk through a concrete example.

15. Dynamo uses vector clocks. Does that mean Dynamo provides causal consistency? Explain the distinction.

16. You are building a collaborative document editor. Two users edit the same paragraph simultaneously. What consistency model do you need and why?

17. What is the cost of causal consistency compared to eventual consistency? What coordination does it require?

---

### Strong Consistency (Linearizability)

18. You are designing a flight seat booking system. Two users try to book the last seat simultaneously. What consistency model do you need and why? What happens if you use eventual consistency instead?

19. Walk me through how Raft achieves strong consistency. Why can't a follower serve reads directly?

20. What is the difference between strong consistency and single-leader synchronous replication? Which is stronger?

21. A payment system processes a debit on one node and a credit on another. What problem arises and how does two-phase commit solve it? What is the risk?

22. Your strongly consistent system is in a network partition. Half the nodes can't reach the other half. What happens and why?

---

### Choosing the Right Model

23. You are designing a shopping cart. Which consistency model and why?

24. You are designing a bank account balance. Which consistency model and why?

25. You are designing a live comment section for a sports event. Comments must appear in order within a thread. Which consistency model and why?

26. You are designing a user session store (login tokens). A user logs out. What happens if another request hits a stale replica that still shows them as logged in? Which model do you need?

27. You are designing an inventory system for concert tickets. 500 tickets, 10,000 concurrent users at sale time. Which consistency model? What breaks if you choose wrong?

28. You are designing a leaderboard for a mobile game that updates every 5 minutes. Which model and why?

---

### CAP Theorem & Partitions

29. Explain CAP theorem in one sentence without using the words "consistency", "availability", or "partition tolerance."

30. A network partition splits your cluster into two halves. You have two choices: both halves keep serving requests (possibly with stale data), or one half stops serving. Which choice corresponds to AP and which to CP?

31. Why is "CA without P" not a realistic option for distributed systems at scale?

32. You chose AP for your system. A network partition heals after 30 seconds. What state is your system in and what needs to happen?
