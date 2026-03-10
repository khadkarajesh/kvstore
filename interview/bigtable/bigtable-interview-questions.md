## Bigtable Paper — Interview Questions

These are open-ended questions as they would appear in a real system design interview.
Answer each out loud or in writing before checking your notes.

---

### Business Motivation

1. Why did Google build Bigtable instead of using GFS or a relational database? What specific gap does it fill?

2. Bigtable serves Gmail (latency-sensitive) and web crawl (throughput-sensitive) on the same system. What design decisions make that possible?

3. What does "structured data at Google scale" mean? Why can't SQL handle it?

---

### Data Model

4. Walk me through Bigtable's data model — row, column family, column, timestamp, cell. How do they relate?

5. What is the difference between a column family and a column? Why does this distinction matter for storage?

6. Two rows in the same table have completely different columns. How is this possible and what problem does it solve?

7. How does Bigtable handle versioning? How does garbage collection work without application code?

8. What does single-row atomicity mean? What does Bigtable NOT guarantee across rows?

---

### Row Key Design

9. You are designing Gmail's inbox. What is your row key and why? What access pattern does it optimize for?

10. Why does the web crawl table store row keys as reversed domain names (com.cnn.www instead of www.cnn.com)?

11. A poorly designed row key can make Bigtable perform badly. Give an example of a bad row key and explain why it's bad.

---

### Architecture

12. The master server is not on the read/write path. How does a client find the right tablet server? Walk through the full lookup flow.

13. What are the three systems in Bigtable and what is the one job of each?

14. What is Chubby and what would break if Chubby went down?

15. What is stored in the Chubby namespace? Be specific.

---

### Failures

16. A tablet server dies. Walk through what happens step by step — from death to data being served again.

17. The master dies. Walk through what happens step by step — from death to a new master serving.

18. How does master rebuild its state when it starts up? What sources does it use?

19. A tablet server's Chubby session expires but the server is actually still alive (network hiccup). What happens and why?

---

### Storage Engine

20. What is the write path in Bigtable from client request to durable storage?

21. What are the three types of compaction? When does a deleted record actually get removed from disk?

22. How does a read find data that spans multiple SSTables? What optimization avoids scanning all of them?

23. Why are SSTables immutable? What does that enable?

---

### Refinements

24. What is a locality group? Give a concrete example of when it improves performance.

25. When would you mark a locality group as in-memory? What is the tradeoff?

26. A bloom filter says a key is not in an SSTable. Can you trust that? What if it says the key is present?

27. A read misses the Scan Cache but hits the Block Cache. Walk through what happens.

---

### Trade-offs & Comparisons

28. Bigtable is CP, Dynamo is AP. What does that mean concretely when a node fails in each system?

29. You need to build a time-series metrics system storing CPU readings per server. Would you use Bigtable or Cassandra? Why? What row key would you use?

30. When would you NOT use Bigtable? Give three scenarios with alternatives.

31. An interviewer asks: "Design a system to store and query Google's entire web crawl index." Walk through your design using Bigtable's principles.
