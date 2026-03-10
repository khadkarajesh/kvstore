## Back-of-Envelope Estimation — Interview Questions

These are open-ended questions as they would appear in a real system design interview.
Answer each out loud or in writing before checking your notes.

---

### Reference Numbers

1. A request goes from client to a server in the same datacenter. A request goes from client to a server in another region. What is the approximate latency for each, and what is the ratio between them?

2. You are choosing between serving data from RAM versus from SSD. What is the approximate latency of each, and what is the ratio? When does this choice change your architecture?

3. What is the approximate size of: a single tweet, a photo, one minute of video? Why do you need these numbers memorized before a design interview?

4. How many seconds are in a day? Why do engineers round this to 100,000 and not 86,400?

5. An interviewer asks you to estimate how many Google searches happen per second worldwide. What is your starting point and how do you arrive at a number?

---

### QPS Estimation

6. Twitter has 300M DAU. Each user reads a timeline 10 times per day and tweets once per day. Compute average read QPS, average write QPS, and peak QPS for both. What does the read-to-write ratio tell you about the architecture?

7. Instagram has 500M DAU. Each user views 20 photos per day and uploads 1 photo. The average photo is 200KB. Compute: average read QPS, average write QPS, peak QPS, and ingress bandwidth. What architectural decision does the bandwidth number force?

8. A startup has 1M DAU. An investor asks "can your single PostgreSQL server handle this?" Walk through a QPS estimate to give a data-driven answer.

9. You are designing a URL shortener. Assume a 100:1 read-to-write ratio, 500M new URLs created per month. Compute: write QPS, read QPS, and whether a single server can handle the read load.

10. A payment system processes 1M transactions per day with an average of $100 each. Estimate: transactions per second (average and peak), and storage per year assuming 1KB per transaction record with 3× replication.

---

### Storage Estimation

11. Twitter generates 500M tweets per day. Each tweet is 280 bytes. Compute: storage per day, storage per year, and total storage for 5 years with 3× replication. What type of storage does this number suggest (relational DB, blob store, something else)?

12. Twitter has 50M photos per day (10% of tweets include a photo). Each photo is 200KB after compression. Compute: daily storage, 5-year storage with 3× replication. How does this compare to the text-only tweet storage?

13. A logging system ingests 100,000 log lines per second. Each log line is 500 bytes. Compute: ingress rate in MB/sec, daily storage, and 30-day retention storage. Does this fit on one machine?

14. YouTube has 500 hours of video uploaded per minute. One minute of video is 50MB. Compute: ingress bandwidth in GB/sec. What infrastructure must exist to handle this?

15. A chat application retains messages for 10 years. 100M DAU, each user sends 50 messages per day, average 100 bytes per message. Compute: total messages per day, storage per day, 10-year total with 3× replication.

---

### Bandwidth Estimation

16. What is the difference between ingress and egress bandwidth? For a read-heavy system like Twitter, which typically dominates and why?

17. Twitter serves 30,000 read QPS for timeline requests. The average timeline response is 10KB (text only, no images). Compute egress bandwidth in MB/sec. Now recompute assuming each response includes 5 thumbnail images at 20KB each.

18. A video streaming service has 5M concurrent viewers, each watching at 4Mbps (1080p). Compute the total egress bandwidth required. How many servers do you need if each server handles 10Gbps of egress?

19. An API serves 10,000 QPS. The average response is 2KB. Compute: egress bandwidth per second and per day. Is a single NIC sufficient or do you need multiple servers?

20. A CDN serves 80% of your traffic. Your origin bandwidth is 10GB/sec without the CDN. What is the approximate origin bandwidth with the CDN? What does this mean for the number of origin servers required?

---

### Full System Estimates

21. "Design WhatsApp." Before touching architecture, what three numbers do you estimate first, and how do those numbers drive your first architectural decisions?

22. You are estimating storage for a distributed key-value store serving 100K writes per second. Each value is 1KB. How much new data is written per day? Per year? With 3× replication, how many storage nodes do you need assuming each node has 10TB of usable storage?

23. A search index for an e-commerce site has 100M product listings, each with 5KB of metadata. How large is the full index? If you cache the top 20% (by traffic), how much RAM do you need for the cache? Does this fit on one Redis node?

24. Walk through a complete estimation for "Design a rate limiter storing per-user counters." Assume 50M active users, each requiring 2 counters of 8 bytes. Can the entire counter set fit in Redis on a single node? What does this mean for the architecture?

25. An interviewer gives you no numbers and says "design Twitter." Walk through the estimation phase: what assumptions do you state, what four numbers do you compute, and what single sentence connects those numbers to your design choices?
