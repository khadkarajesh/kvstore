## Search Systems — Flashcards

Format: Q → A
Cover the answer, recall it, then check.

---

**Q: What is an inverted index and how does it differ from a forward index?**
A: Forward index: document → list of words it contains. Inverted index: word → list of documents containing it. Inverted index enables O(1) term lookup vs O(N) full scan with a forward index. It is the core data structure behind every search engine.

---

**Q: Walk through the four steps of building an inverted index from raw text.**
A: 1) Tokenization — split text into tokens. 2) Analysis — lowercase, remove stop words, apply stemming/synonyms. 3) Posting list construction — for each token record (doc_id, term_frequency, positions). 4) Sort and merge — sort posting lists by doc_id to enable fast intersection for multi-term queries.

---

**Q: What is gap encoding (delta compression) in posting lists and why does it help?**
A: Instead of storing absolute doc IDs [1, 5, 10, 12], store deltas [1, 4, 5, 2]. Deltas are small integers → compress better with variable-length encoding. Reduces disk and memory footprint of posting lists significantly.

---

**Q: What is tokenization in Elasticsearch's analysis chain?**
A: Splitting raw text into individual tokens. "The Quick Brown Fox!" → ["The", "Quick", "Brown", "Fox"]. Tokenization happens at both index time (when building the inverted index) and query time (so both use the same representation). Inconsistent analysis between index and query time causes missed results.

---

**Q: What is TF-IDF and what are the two components?**
A: TF (Term Frequency): how often a term appears in a document — more occurrences = higher score. IDF (Inverse Document Frequency): log(N / df) — how rare a term is across all documents. Rare terms score higher because they are more informative. Final score = TF * IDF.

---

**Q: What is the key weakness of TF-IDF that BM25 fixes?**
A: TF does not saturate in TF-IDF — doubling occurrences doubles the score. A long document with a term repeated 50 times scores higher than a short document where the term appears 5 times, even if the short document is more relevant. BM25 adds a saturation parameter k1 and length normalization parameter b to fix this.

---

**Q: In BM25, what do the parameters k1 and b control?**
A: k1 controls term frequency saturation — typically 1.2 to 2.0. Higher k1 = more weight given to raw term count before plateauing. b controls length normalization — b=1 fully penalizes long documents, b=0 disables it. Default in Elasticsearch: k1=1.2, b=0.75.

---

**Q: What is an Elasticsearch shard and why does the primary shard count matter?**
A: A shard is a single Lucene instance — the unit of storage and distribution. Primary shards receive writes; replicas serve reads. The primary shard count is fixed at index creation time and cannot be changed without a full reindex. Over-provision primary shards when creating an index if growth is expected.

---

**Q: What is an Elasticsearch replica and what two purposes does it serve?**
A: A replica is a complete copy of a primary shard. Purpose 1: fault tolerance — if the primary's node dies, a replica is promoted to primary with no data loss. Purpose 2: read scaling — replicas serve search queries, so adding replicas increases read throughput linearly.

---

**Q: Walk through Elasticsearch's scatter-gather read path.**
A: 1) Coordinating node receives query. 2) Scatter: sends query to one copy of each shard. 3) Each shard runs query locally, returns top-K doc IDs + scores (not full docs). 4) Coordinating node merges and re-ranks → selects global top-K. 5) Fetch: coordinating node retrieves full documents from relevant shards. 6) Returns assembled result to client.

---

**Q: Why is Elasticsearch called "near-real-time" instead of real-time?**
A: Writes go to an in-memory buffer and the translog (WAL). The buffer is flushed to a searchable Lucene segment only during a refresh (default: every 1 second). A document is not searchable until the segment is written and opened. Writes are durable (translog) but not immediately visible in search results.

---

**Q: What is a Lucene segment and what happens when there are too many of them?**
A: A segment is an immutable on-disk file containing a portion of the inverted index. Created on each refresh. More segments = more files to scan per query = slower searches. A background merge process combines small segments into larger ones. During bulk indexing, disable refresh and merge to avoid overhead, then enable after load completes.

---

**Q: What is the write path for a document from client to Elasticsearch primary shard?**
A: 1) Coordinating node receives write. 2) Routes to correct primary: shard = hash(doc_id) mod num_primary_shards. 3) Primary runs analysis pipeline on document. 4) Writes to in-memory buffer and translog (flushed to disk). 5) Forwards to all replicas in parallel. 6) Acks client after replicas confirm. Document becomes searchable on next refresh.

---

**Q: What is custom routing in Elasticsearch and when do you use it?**
A: PUT /index/_doc/42?routing=user_id_789 forces a document to land on a specific shard based on the routing key. When querying with the same routing key, only one shard is searched instead of all shards. Eliminates scatter-gather overhead. Use case: per-user data where all of a user's documents should co-locate on one shard.

---

**Q: What is a trie and how does it support autocomplete?**
A: A tree where each node is a character; paths from root to terminal nodes spell complete words. Prefix lookup: walk the trie matching each character → O(L) where L = prefix length. All completions reachable via DFS from the prefix node. Rankings stored as scores at terminal nodes. Memory intensive for large vocabularies; hard to distribute.

---

**Q: What is the Elasticsearch Completion Suggester and how is it different from a prefix query?**
A: Completion Suggester uses an FST (Finite State Transducer) stored in-memory, purpose-built for prefix lookups with weights. Much faster than a prefix query (which scans the inverted index for all matching terms). Prefix query works for low-cardinality fields; Completion Suggester is used for production autocomplete at scale. Trade-off: Completion Suggester is less flexible and requires a dedicated mapping field.

---

**Q: What is edge n-gram tokenization and why is it used for autocomplete?**
A: At index time, "search" is tokenized into ["s", "se", "sea", "sear", "searc", "search"]. A query for the prefix "sea" then matches the "sea" token directly via normal term lookup — no special prefix query needed. More flexible than Completion Suggester but inflates index size by 5–10x for the tokenized field.

---

**Q: Elasticsearch master node — what does it do and how is split-brain prevented?**
A: Master node manages cluster state: creating/deleting indices, adding/removing nodes, assigning shards. Split-brain prevention: quorum-based election — a partition must have a majority of master-eligible nodes to elect a master. Setting in ES <7: minimum_master_nodes = (N/2)+1. ES 7+ uses automated voting configuration that eliminates manual quorum configuration.

---

**Q: What happens in Elasticsearch when a primary shard's node dies?**
A: 1) Master detects node failure (within election timeout, ~30s). 2) Master promotes the most up-to-date replica to new primary (tracked by sequence numbers). 3) New replica assigned to a live node to restore redundancy. 4) Cluster moves from green → yellow (during promotion) → green (when new replica is allocated). In-flight writes not yet replicated to the replica are lost.

---

**Q: When would you use Postgres full-text search (GIN index) instead of Elasticsearch?**
A: When dataset fits on a single DB instance (<10M documents), search is a secondary feature (not core product), the team wants to avoid operating a separate Elasticsearch cluster, and relevance requirements are simple (basic ranking via ts_rank is sufficient). LIKE '%term%' is never acceptable in production — always use GIN-indexed tsvector for Postgres full-text.

---

**Q: What is write amplification in Elasticsearch?**
A: A single document write updates multiple internal structures: inverted index, stored source (_source), doc values (for sort/aggregations), BKD tree (for numeric range queries), and is replicated to all replica shards. Total write cost is typically 5–10x raw document size. Use the _bulk API to amortize this overhead during high-volume ingestion.

---

**Q: What are doc values in Elasticsearch and why do they exist?**
A: Doc values are a columnar, on-disk data structure that stores field values sorted by document ID. Inverted index is optimized for term → doc lookup (search). Doc values are optimized for doc → field value lookup (sorting, aggregations, scripting). Without doc values, sorting and aggregations would require loading all field values into heap memory.

---

**Q: Why is Elasticsearch not a system of record?**
A: Elasticsearch does not provide ACID transactions or strong consistency guarantees. Writes are near-real-time (not immediately searchable). After node failure, in-flight writes may be lost. Index state is eventually consistent across the cluster. It should be populated from a durable source of truth (e.g. Postgres) via a sync pipeline, not used as the primary database.

---

**Q: What are the two indices Twitter maintains for search and why?**
A: 1) Real-time index: recent tweets (last 7 days), kept in-memory for sub-second latency, indexed within seconds of posting. 2) Archive index: full tweet history, stored on disk, slower but comprehensive. Queries are executed against both and results merged by recency and engagement score. Splitting by age bounds real-time index size and latency.

---

**Q: What is segment merging and what is the operational risk during high write load?**
A: Background Elasticsearch/Lucene process that combines many small segments (created by frequent refreshes) into fewer large segments. Fewer segments = faster queries. Risk: merge is CPU and I/O intensive — during high write load, continuous merging competes with indexing and search, causing latency spikes. Mitigation: throttle merge rate, use dedicated data nodes for write-heavy indices, schedule bulk loads during off-peak hours.

---
