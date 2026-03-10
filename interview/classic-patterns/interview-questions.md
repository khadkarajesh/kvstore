Classic System Design Patterns — Interview Questions

These are open-ended questions as they would appear in a real system design interview.
Answer each out loud or in writing before checking your notes.

---

### URL Shortener

1. A user submits the same long URL twice. Should the system return the same short code or generate a new one? What are the tradeoffs of each approach?

2. You chose MD5 hashing to generate short codes. Two different long URLs produce the same 7-character prefix. How do you detect and resolve this collision?

3. Your product team wants short links to redirect with 301. Your analytics team wants to track every click. Why is 301 at odds with analytics, and how do you resolve it?

4. A user wants a custom alias: company.com/launch. How does your system handle alias uniqueness? What happens if the alias is already taken?

5. Your URL shortener has a 100:1 read/write ratio. Where specifically do you add caching, what is your cache key, and how do you handle cache invalidation when a URL expires?

6. A single user is generating 10,000 short URLs per minute via the API. How do you detect and stop this, and where in the stack do you enforce the limit?

---

### News Feed

7. You are designing a news feed for 500M users. Each user has up to 5,000 followers. Walk through the tradeoffs between fan-out on write and fan-out on read.

8. A celebrity has 50 million followers and posts once per hour. If you use fan-out on write, what happens to your write throughput? How do you handle this specifically?

9. Your feed ranking algorithm uses recency, engagement score, and relationship strength. A post's engagement score changes every second as likes come in. How do you keep the ranked feed fresh without re-ranking on every read?

10. A user scrolls to the bottom of their feed and asks for more posts. How does your pagination work? Why is offset-based pagination a bad idea here?

11. How would you store the feed for each user? Walk through your schema — what database, what data model, and why. How long do you retain feed data?

12. A user unfollows someone. Should you immediately purge that person's posts from the feed? What are the tradeoffs of eager vs lazy removal?

---

### Chat System

13. Two users are in a chat. User A sends message M1, then M2. User B receives M2 before M1. How does this happen and how do you guarantee ordering within a conversation?

14. Your chat system has 10 million concurrent WebSocket connections. A message needs to be delivered to a user on a different server than the sender. Walk through your architecture.

15. User B is offline when user A sends a message. When B comes back online 3 days later, how does your system deliver the missed messages? What storage design supports this?

16. A group chat has 5,000 members. A single message must be delivered to all of them. How does fan-out work here — server-side or client-side? What breaks at 5,000 members?

17. How do you show whether a user is online, typing, or last seen? Walk through your presence and typing indicator design from client heartbeat to UI update.

18. An interviewer asks: "How do you do end-to-end encryption architecturally?" You don't need to explain the cryptography — explain what it means for your server's role, storage, and ability to moderate content.

---

### Notification System

19. Your system must deliver notifications via push (APNs/FCM), SMS, and email. A user has all three enabled. How do you decide which channel to use, and in what order?

20. FCM rejects 5% of push notification deliveries with a 503. How does your retry strategy work? How do you avoid hammering FCM with retries while still delivering reliably?

21. A user receives the same "you have a new follower" notification 3 times due to a retry bug. How do you make notification delivery idempotent?

22. Your system sends 10 million notifications per hour during peak. How do you architect the delivery pipeline to avoid overloading downstream providers (APNs, Twilio)?

23. A breaking news event triggers notifications for 50 million users simultaneously. How do you handle this spike without your notification queue becoming a bottleneck?

24. A user opts out of marketing notifications but still wants security alerts. How does your system model notification categories and user preferences?

---

### Web Crawler

25. You are crawling the web starting from 100 seed URLs. Should you use BFS or DFS? Which produces a more useful crawl in practice and why?

26. Your crawler visits the same URL from two different workers simultaneously. How do you deduplicate URLs at scale — what data structure and where does it live?

27. A website's robots.txt says: "Crawl-delay: 10". Your crawler has 100 workers all hitting the same domain. How do you enforce politeness per domain across distributed workers?

28. Your crawler makes a DNS lookup for every URL. At 10,000 URLs/second this becomes a bottleneck. How do you fix it, and what are the consistency risks of caching DNS results?

29. You have 1,000 crawler workers. A single URL frontier (queue) is the bottleneck. How do you shard the frontier across workers while keeping per-domain politeness intact?

30. A page you crawled last month has changed. How does your system decide when to re-crawl a page? Walk through a freshness model that balances recrawl cost against content staleness.
