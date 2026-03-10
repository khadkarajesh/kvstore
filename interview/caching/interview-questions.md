## Caching — Interview Questions

These are open-ended questions as they would appear in a real system design interview.
Answer each out loud or in writing before checking your notes.

---

### Cache Fundamentals

1. Why does caching exist? What three problems does it solve, and which one is the most important at scale?

2. A cache has a 95% hit rate. The DB can handle 10K QPS. The app receives 200K QPS. Is this setup safe? Walk through the math.

3. What is the difference between cache hit rate and cache miss rate? At what hit rate does caching stop providing significant benefit?

4. You restart your app servers and your DB gets hammered for 60 seconds before recovering. What happened and how do you prevent it?

5. What does "cache as a system of record" mean and why is it dangerous?

---

### Cache Strategies

6. Your app has a user profile that is read 10,000 times per hour and updated once per day. Which cache strategy do you use and why?

7. You are building a payment system. A user's balance must never be stale. What cache strategy do you use? What are the tradeoffs?

8. Describe the exact sequence of operations in a cache-aside read that results in a stale read reaching the user.

9. Write-through guarantees consistency but has a specific failure mode. What is it, and when does it appear?

10. Your team is considering write-behind caching to reduce DB write load by 10x. What questions do you ask before approving it?

---

### Redis Architecture

11. Redis is single-threaded yet benchmarks at 100K ops/sec. Explain why the single-threaded model is a performance advantage rather than a limitation.

12. You need to implement a real-time leaderboard for a game with 10 million players. Which Redis data structure do you use and why? What commands?

13. A Redis master dies. Walk through what happens in a Sentinel setup — from death detection to clients being redirected to the new master.

14. You have 6 Redis nodes and want to shard with Redis Cluster. How are keys distributed across nodes? How does a client know which node to query?

15. Your team debates RDB vs AOF for Redis persistence. The use case is session storage — sessions last 24 hours, losing 30 seconds of data is acceptable. Which do you recommend and why?

---

### Cache Problems

16. At 3am your DB CPU spikes to 100% for 90 seconds and then drops back to normal. No deployments happened. What is the most likely cause and how do you confirm it?

17. A bad actor queries your public API with 10 million random non-existent user IDs. Your DB starts receiving 10 million queries per second. What attack is this and what are two defenses?

18. You have a Redis key that holds the Twitter home timeline for a celebrity with 50 million followers. The key gets 500K reads per second. Your Redis shard handling that key is at 100% CPU. What do you do?

19. Your team adds caching to improve read performance. After deploying, read latency improves by 80% but write latency doubles. What cache strategy did they likely implement and is this acceptable?

20. You are using TTL-based cache invalidation with a 5-minute TTL. A user updates their email address. How long can another part of the system see the old email? How do you reduce this window without removing TTL entirely?

---

### System Design Application

21. You are designing Twitter's home timeline. A user follows 500 people. Each of those people tweets up to 100 times per day. Where do you add caching, at what layer, and what eviction policy?

22. You are designing a global e-commerce product catalog with 10 million SKUs. Prices change every few minutes. Product descriptions change rarely. How do you cache differently for prices vs descriptions?

23. You are designing a rate limiter that allows 100 requests per user per minute across a fleet of 50 app servers. Why can't you use a local cache? What do you use instead, and which Redis data structure?

24. You are designing a distributed session store for a web app with 100 million daily active users. Sessions must survive a single node failure. Walk through your caching architecture.

25. An interviewer asks: "Design YouTube's view count system — views must be approximately correct within 60 seconds, and the system handles 1 billion views per day." Walk through your caching and aggregation strategy.
