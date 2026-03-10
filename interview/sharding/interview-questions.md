## Sharding — Interview Questions

These are open-ended questions as they would appear in a real system design interview.
Answer each out loud or in writing before checking your notes.

---

### Sharding Strategies

1. What is sharding and why does it exist? What are the three specific conditions that force you to shard?

2. Walk through range-based sharding with a concrete example. What is its key advantage for range queries, and what is the hotspot failure mode?

3. Walk through hash-based sharding with a concrete example. What is its key advantage over range-based, and what catastrophic problem occurs when you add a new shard?

4. What is directory-based sharding? What flexibility does it offer, and what two operational risks does it introduce?

5. What is consistent hashing? Contrast it with hash(key) % N: when you add one node to a 4-node cluster, what fraction of keys must move with each approach?

6. What are virtual nodes in consistent hashing? What three specific problems do they solve that a simple ring without virtual nodes cannot?

---

### Shard Key Design

7. You are designing a sharded social network. An engineer proposes sharding the Posts table by `post_id`. A second engineer proposes sharding by `user_id`. Which is better and why? What query pattern does each optimize?

8. What is the fundamental design principle for choosing a shard key? State it in one sentence, then apply it to an e-commerce system with orders, customers, and products.

9. You are sharding a time-series database (CPU metrics per server, timestamped). An engineer proposes using `timestamp` as the shard key so you can do time-range queries on a single shard. What is the severe problem with this choice?

10. How do you fix a sequential auto-increment ID as a shard key in a range-sharded system? Walk through at least two techniques.

11. You are designing a sharded URL shortener. The short code is 7 characters (random). What is your shard key and why? Are range queries ever needed, and does that affect your choice?

---

### Hotspots & Rebalancing

12. What is a hot shard? Give three different causes of hot shards — one from key distribution, one from access patterns, and one from temporal patterns.

13. Taylor Swift tweets. Her tweet is accessed 100 million times in an hour while the average tweet is accessed 100 times. How does key prefix salting fix this, and what is the write amplification cost?

14. Walk through how you detect a hot shard in production. What metrics indicate a hotspot, and what is the first operational action you take?

15. You have a hash-sharded database with 4 shards. You need to add a 5th shard. Walk through the steps of the resharding migration without downtime. What is dual-write and when do you stop it?

16. What is pre-sharding? Why is it easier to start with 1,024 virtual shards mapped to 4 physical nodes than to start with 4 logical shards?

17. An interviewer asks: "Your shard is running hot. What are your options?" Give a structured answer with at least four distinct techniques, from cheapest to most operationally expensive.

---

### Cross-shard Operations

18. A SQL query joins the Users table (sharded by user_id) and the Orders table (sharded by order_id). What problem does this create, and what is the name of the operation required to execute it?

19. What is scatter-gather? What is the latency cost of a scatter-gather query across 100 shards versus a single-shard query?

20. You need a distributed transaction across two shards (debit from shard 1, credit to shard 2). Walk through Two-Phase Commit across shards. What happens if the coordinator crashes between phases?

21. The production guidance is "avoid 2PC across shards." What is the primary design technique to eliminate the need for cross-shard transactions?

22. An interviewer asks: "Design a friends-of-friends query for a social network sharded by user_id." Walk through why this query is hard to execute efficiently and what trade-offs you accept.

---

### Real-world Design

23. You are designing Twitter's tweet storage at scale (500M tweets/day, 5-year retention). Walk through: why you shard, which strategy you use, what your shard key is, and how many shards you estimate you need.

24. A database has 10 shards, each with 3 replicas (primary + 2 followers). Total: 30 nodes. One primary shard fails. Walk through what happens: which data is affected, how long is the outage, and what data is at risk with async replication?

25. An interviewer says "how do you scale a database beyond what a single machine can handle?" Give a complete structured answer: vertical scaling first, then read replicas, then sharding — including when you make each transition and what shard key you would choose for a social media platform.
