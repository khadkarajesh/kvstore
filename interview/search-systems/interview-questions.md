## Search Systems — Interview Questions

These are open-ended questions as they would appear in a real system design interview.
Answer each out loud or in writing before checking your notes.

---

### Inverted Index Fundamentals

1. A product catalog has 50 million items. Explain why a full-table scan with LIKE '%running shoe%' is unacceptable in production and what data structure makes keyword search fast. Walk through how a query for "running shoe" executes against that structure.

2. You are building an inverted index from scratch. A document contains the sentence "The Quick Brown Fox Jumps." Walk through every step from raw text to a posting list entry, naming each stage of the analysis pipeline and what it does.

3. Two documents both contain the word "database" ten times. Document A is 100 words long; Document B is 10,000 words long. Under TF-IDF, which one scores higher? Under BM25 with default settings, which one scores higher? Explain why BM25's answer is more useful.

4. Your search index has 1 billion documents. The term "the" appears in 999 million of them. The term "linearizable" appears in 3,000. A user searches for "linearizable the distributed systems." Explain how IDF affects which terms drive the final ranking and why stop word removal is useful.

5. You need to support multi-word phrase search — the user wants documents where "distributed systems" appears as an exact phrase, not just documents containing both words independently. What additional information must the inverted index store beyond doc IDs and term frequency?

---

### Elasticsearch Architecture

6. Your team is launching a new product search feature. The product catalog has 20 million documents today and is expected to grow to 200 million over two years. How do you choose the number of primary shards at index creation time? What happens if you get it wrong?

7. You have a 6-node Elasticsearch cluster. An index has 3 primary shards and 2 replicas per shard. Draw the shard assignment. Which single node failure can you survive? What is the minimum number of nodes you need for this configuration to be fully resilient to any single node failure?

8. A user searches for a product. The search returns 0 results even though the product was indexed 300ms ago. The same query returns the product 2 seconds later. Explain exactly what happened at each layer of Elasticsearch's write path to produce this behavior.

9. Your team wants to disable Elasticsearch's translog to improve write throughput during a one-time data migration of 500 million documents. Is this safe? What is the translog's purpose and what do you lose by disabling it?

10. Elasticsearch's coordinating node is becoming a CPU bottleneck — merging results from 50 shards for complex aggregation queries. What is the architectural cause of this and what are two ways to address it?

---

### Relevance and Ranking

11. Your e-commerce search returns highly relevant results for exact product names but ranks promotional blog posts above products when users search for generic terms like "running shoes." What mechanism in Elasticsearch lets you weight the title field more heavily than the description field, and what query type do you use?

12. A user searches for "shoe." Your index contains documents using "shoes" (plural), "shoeing" (verb), and "shoemaker" (compound). Which of these should match the query and which should not? What analysis component controls this, and what is the difference between stemming and lemmatization?

13. Your team wants to incorporate user engagement signals (click-through rate, purchase rate) into the search ranking rather than relying purely on BM25 relevance. What Elasticsearch mechanism allows you to blend a text relevance score with a pre-computed numeric engagement score? Describe the query structure.

14. A search for "Apple" should return Apple Inc. products ahead of apple (fruit) recipes on a tech product website. BM25 alone cannot distinguish this because it has no domain context. What two approaches can you use in Elasticsearch to inject this context-awareness into ranking?

15. You are running A/B tests on two ranking models. Model A uses BM25 defaults. Model B uses a custom function score that boosts by recency and CTR. Both serve 50% of traffic. Describe the infrastructure changes needed to serve both models simultaneously from a single Elasticsearch cluster without one model contaminating the other.

---

### Autocomplete / Typeahead Design

16. Explain the difference between using an Elasticsearch prefix query and using the Completion Suggester for autocomplete. When does the prefix query become a performance problem, and what is the Completion Suggester's core data structure that makes it faster?

17. You are building autocomplete for a search bar that receives 50,000 requests per second at peak. The suggestions are derived from the top 10 million most frequent historical search queries. A trie built from these queries fits in 4GB of RAM. Design the serving architecture — where does the trie live, how do you handle updates, and what is your fallback for rare prefixes not covered by the trie?

18. Google autocomplete shows personalized results — if you have searched for "python tutorial" before, typing "py" suggests "python" ahead of "physics." Design the data flow from user query log → personalized suggestion model → serving layer. What latency budget does each layer get?

19. Your autocomplete returns suggestions in 80ms at p50 but 600ms at p99. After profiling, the bottleneck is the Elasticsearch Completion Suggester under high concurrency. List three specific changes you would make to the architecture to reduce p99 latency to under 150ms.

20. A competitor can find a user's partial searches by probing your autocomplete API (type "j", "jo", "joh", "john" and see which completions appear). This leaks the fact that someone searched for a specific sensitive query. How do you design an autocomplete system that provides useful suggestions while preventing this kind of privacy inference attack?

---

### Real-World Design Problems

21. Design Twitter Search. Twitter ingests 500 million tweets per day. Users expect search results to include tweets posted in the last 30 seconds. Walk through the full system: ingestion pipeline, index architecture, query path, ranking signals, and how you handle the freshness requirement. What are the two hardest scaling problems?

22. Design Google Autocomplete. 8 billion searches per day. Autocomplete must respond in under 100ms globally. Completions should reflect trending queries within the last hour. Walk through the offline pipeline, the serving layer, caching strategy, and how you handle real-time trending (e.g. a breaking news event causes a new query to spike to top-10 globally within minutes).

23. Add search to an e-commerce platform with 50 million products, 10 million daily active users, and a peak query rate of 100,000 searches per second. The team currently uses Postgres for product data. Walk through the full architecture: how data gets from Postgres into the search index, how you shard and scale Elasticsearch, how you handle faceted search (filter by category + brand + price), and what SLAs you target.

24. Design a log analytics system that ingests 1 TB of logs per day and must support full-text search across all logs from the last 30 days with query latency under 2 seconds. Users frequently search for error messages with specific stack trace fragments. Walk through the ingestion pipeline, index lifecycle management (hot/warm/cold tiers), and how you keep costs manageable at scale.

25. You are the search infrastructure lead at a company. The product team wants to add semantic search — "find products similar in meaning to my query even if they don't share exact keywords." For example, searching "comfortable work shoes" should match products described as "ergonomic office footwear" even with no keyword overlap. The current system is Elasticsearch-based. Walk through the technical approach: what additional components you need, how you would integrate them with the existing keyword search, and what the deployment and operational complexity looks like.
