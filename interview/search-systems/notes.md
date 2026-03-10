Search Systems — System Design Notes

---
Inverted Index: What It Is and How It's Built

The core data structure behind every search engine.

Forward Index (what a database naturally has):
  Document ID → list of words in that document
  doc1 → ["the", "quick", "brown", "fox"]
  doc2 → ["the", "fox", "jumped", "high"]
  Querying "fox": must scan every document and check if "fox" is present → O(N) full scan

Inverted Index (what search engines build):
  Word → list of (document ID, position) pairs
  "fox"     → [(doc1, pos=4), (doc2, pos=2)]
  "quick"   → [(doc1, pos=2)]
  "jumped"  → [(doc2, pos=3)]
  Querying "fox": look up "fox" in index → O(1) lookup → get matching doc list immediately

Building an Inverted Index:
  Step 1 — Tokenization: split raw text into tokens
    "The Quick Brown Fox!" → ["The", "Quick", "Brown", "Fox"]
  Step 2 — Normalization / Analysis:
    Lowercasing:       "The" → "the"
    Stop word removal: "the" dropped (too common to be useful)
    Stemming:          "running" → "run", "jumped" → "jump"
    Synonyms:          "automobile" → also index as "car"
  Step 3 — Posting list construction:
    For each token, record (doc_id, term_frequency, positions)
  Step 4 — Sort and merge:
    Posting lists sorted by doc_id for efficient intersection (AND queries)
    Multi-word query "brown fox" → intersect posting list of "brown" with posting list of "fox"

Storage:
  In-memory: hash map of term → compressed posting list
  On-disk:   posting lists serialized, term → byte offset in file
  Compression: doc IDs stored as deltas (gap encoding) — store diff between consecutive IDs, not absolute IDs
    [1, 5, 10, 12] → [1, 4, 5, 2] → more compressible

---
Elasticsearch Architecture

Elasticsearch is a distributed search engine built on top of Apache Lucene.
Lucene handles the actual index logic (inverted index, scoring).
Elasticsearch adds distribution, replication, and a REST API on top.

Key Concepts:

  Cluster: a group of nodes that collectively hold all data and provide search
  Node: a single Elasticsearch server process; can play different roles
    Master node:     cluster coordination — creating/deleting indices, adding nodes
    Data node:       stores shards; executes search and indexing requests
    Coordinating node: routes requests, merges results (any node can do this)
    Ingest node:     pre-processes documents before indexing (transforms, enrichment)

  Index: analogous to a database table; a logical namespace for documents
    An index is split across multiple shards

  Document: the unit of data — a JSON object with fields
    { "id": "42", "title": "Nike Air Max", "description": "Running shoe", "price": 129.99 }

  Shard: a single Lucene instance; the unit of distribution
    Primary shard: the original copy — receives writes
    Replica shard: a copy of a primary; serves reads; promoted on primary failure
    Rule: number of primary shards is fixed at index creation time (cannot be changed without reindex)
    Typical sizing: 10-50GB per shard; more shards = more parallelism but more overhead

  Example cluster layout (3 nodes, index with 3 primary shards, 1 replica each):
    Node 1: primary-0, replica-1
    Node 2: primary-1, replica-2
    Node 3: primary-2, replica-0
    → each node stores 2 shards; no single node loss causes data loss

---
Write Path: Document Ingestion → Index Update

Client sends PUT /index/_doc/42 with JSON body.

Step 1 — Coordinating node receives the request.
  Determines which primary shard owns doc 42:
    shard = hash(doc_id) mod num_primary_shards
  Routes request to that primary shard's node.

Step 2 — Primary shard processes the write:
  Passes document through the analysis chain (tokenize, normalize, stem).
  Writes to the in-memory buffer (not yet searchable).
  Writes to the translog (transaction log on disk — crash recovery, equivalent to WAL).
  Acks the write once translog is flushed (durable).

Step 3 — Replication:
  Primary forwards the indexed document to all replica shards in parallel.
  Waits for acknowledgment from replicas (configurable: wait_for_active_shards).
  Returns success to client.

Step 4 — Segment refresh (near-real-time):
  Periodically (default: every 1 second), the in-memory buffer is written to a new Lucene segment on disk.
  The segment becomes searchable only after the refresh.
  → This is why Elasticsearch is "near-real-time" — writes are not immediately searchable.
  Forcing: POST /index/_refresh (expensive; avoid calling per-document in production)

Step 5 — Segment merge (background):
  Lucene creates many small segments; background thread merges them into larger ones.
  Smaller segment count = faster queries (fewer files to scan).
  Merge is CPU and I/O intensive; tune merge policy to avoid impacting search latency.

---
Read Path: Query → Results

Client sends GET /index/_search with query body.

Step 1 — Coordinating node receives the request.
  Determines which shards (primary or replica) hold relevant data.
  Scatter: sends query to one copy of each shard (round-robin across primary + replicas for load balancing).

Step 2 — Each shard executes locally (Query phase):
  Runs the query against its local Lucene index.
  Computes relevance score for each matching document.
  Returns top-K document IDs and scores to the coordinating node.
  Does NOT send full documents yet (reduces network transfer).

Step 3 — Coordinating node merges (Gather phase):
  Receives top-K lists from all shards.
  Merges and re-ranks by score.
  Selects global top-K documents.

Step 4 — Fetch phase:
  Coordinating node asks each relevant shard for the full document source.
  Shards return full document JSON.
  Coordinating node assembles final response and returns to client.

This two-phase design (query then fetch) avoids transferring full documents during the merge step.

---
Near-Real-Time Search

Why there is a delay:
  Lucene writes to an in-memory buffer.
  The buffer is only flushed to a disk segment on refresh.
  A segment is only searchable after it is written to disk and opened by a reader.
  Default refresh interval: 1 second → writes are visible within ~1 second.

Controlling the delay:
  Increase refresh interval:   PUT /index/_settings { "refresh_interval": "30s" }
    → better write throughput (fewer segments created), worse freshness
  Decrease refresh interval:   "refresh_interval": "100ms"
    → faster visibility, more segment overhead
  Disable during bulk load:    "refresh_interval": "-1" → manual refresh after bulk insert
  Translog vs segment:         translog (WAL) is written synchronously on every write for crash safety;
                               the segment refresh is what makes data searchable

Practical guidance:
  For product search or log analytics: 1s default is fine.
  For real-time chat or metrics: consider reducing to 100ms or use a different tool.
  During initial data load (bulk indexing): disable refresh, then trigger once at the end.

---
Autocomplete / Typeahead

Problem: given a partial string typed by the user, return relevant completions in <100ms.

Trie (Prefix Tree):
  Structure: tree where each node is a character; paths spell words
    "car" → c → a → r → (end)
    "card" → c → a → r → d → (end)
    "care" → c → a → r → e → (end)
  Prefix lookup: walk trie matching each character → O(L) where L = prefix length
  Finding all completions: DFS from prefix node → collect all terminal nodes
  Ranking: store frequency/score at each terminal node; return top-K by score
  Weakness: memory intensive for large vocabularies; hard to distribute across nodes

Prefix Index (Elasticsearch approach):
  Store n-gram tokens for each searchable term during indexing
    "search" → ["s", "se", "sea", "sear", "searc", "search"]
  Query for prefix "sea" → matches "search" because "sea" is in its token list
  Alternative: edge n-gram tokenizer (common Elasticsearch pattern)

Elasticsearch Completion Suggester:
  Specialized data structure (FST — Finite State Transducer) optimized for prefix lookup
  Stored separately from the main inverted index; kept in-memory for speed
  Supports: fuzzy matching, geographic filtering, context filtering (e.g. category)
  Usage: suggest field with type: completion in mapping
  Returns completions ranked by a pre-computed weight field
  Limitation: not as flexible as full-text search; updating suggestions requires reindexing

Elasticsearch Prefix Query:
  { "prefix": { "title": "sear" } }
  Scans the inverted index for all terms starting with "sear"
  Works fine for small cardinality fields; slow for fields with millions of unique terms
  Best combined with edge n-gram for production at scale

Design for 10M QPS autocomplete (Google scale):
  Cache top completions per prefix in Redis (key = prefix, value = JSON completion list)
  Build trie offline (MapReduce/Spark) from query logs; push to Redis on schedule
  Real-time component: Elasticsearch handles long-tail prefixes not in cache
  Tiered: L1 = local in-process cache for 3-char prefixes, L2 = Redis, L3 = Elasticsearch

---
Relevance Scoring: TF-IDF and BM25

TF-IDF (Term Frequency — Inverse Document Frequency):

  TF (Term Frequency): how often does the term appear in this document?
    TF = count(term, doc) / total_terms(doc)
    More occurrences → higher score. Diminishing: raw count is usually sqrt-compressed.

  IDF (Inverse Document Frequency): how rare is this term across all documents?
    IDF = log(N / df(term))   where N = total docs, df = docs containing the term
    Rare term "linearizable" → high IDF (informative)
    Common term "the" → low IDF (not informative) → naturally down-weighted

  Final score: TF-IDF = TF(term, doc) * IDF(term)

  Weakness: TF does not saturate — doubling term frequency doubles score
    A 1000-word document with "fox" appearing 50 times scores higher than a 50-word document
    where "fox" appears 5 times, even though the shorter doc is arguably more relevant.

BM25 (Best Match 25) — Elasticsearch default since version 5:

  Improves on TF-IDF with two parameters:
    k1 (term saturation): controls how quickly TF saturates (typically 1.2–2.0)
      At k1=1.2: doubling occurrences does NOT double the score — score plateaus
    b (length normalization): controls penalty for long documents (typically 0.75)
      b=1: full normalization — long docs penalized
      b=0: no length normalization

  Formula (simplified):
    score = IDF * (TF * (k1 + 1)) / (TF + k1 * (1 - b + b * docLen/avgDocLen))

  Result: short documents with the search term score higher than long documents with many
  occurrences but lower term density. More intuitive relevance for most use cases.

Boosting and Field Weighting:
  Not all fields are equal: title match > description match > body match
  Boost at query time:   { "match": { "title": { "query": "fox", "boost": 3.0 } } }
  Boost at index time:   set boost in mapping (applied during indexing, not recommended — hard to change)
  Multi-match query: search across multiple fields with different weights:
    { "multi_match": { "query": "fox", "fields": ["title^3", "description^1", "body^0.5"] } }
  Function score: boost by recency, popularity, click-through rate, geographic proximity

---
Scaling

Horizontal shard scaling:
  More primary shards → more parallelism → higher throughput
  Shard count is fixed at index creation → over-provision if growth is expected
  Common pattern: create index with 5-10 primary shards even if starting small
  Re-sharding requires reindex (expensive) → split index API in ES 6+ for doubling

Replica reads:
  Replicas serve read traffic → scale read throughput linearly with replica count
  ES routes searches to replicas in round-robin by default
  Adding replicas: zero downtime, can be done dynamically
  Cost: each replica = full copy of index → 2 replicas = 3x storage

Routing by document ID:
  Default routing: shard = hash(doc_id) mod num_primary_shards
  Custom routing: PUT /index/_doc/42?routing=user_id_789
    → all docs for user 789 land on same shard → queries with routing=user_id_789 hit one shard
    → eliminates scatter-gather for per-user queries; improves query latency

Coordinating node bottleneck:
  All shard results funnel through coordinating node → can be CPU bottleneck for large result merges
  Mitigation: dedicated coordinating nodes (data=false, master=false)

---
Failures

Shard unavailability:
  If a primary shard's node dies:
    Cluster detects node failure (within master election timeout, typically 30s)
    Master elects a replica to become the new primary
    New replica assigned to a live node
    Index is yellow (some replicas unassigned) until a new replica is built
    Cluster is green only when all primaries AND all replicas are assigned

Split-brain prevention:
  Elasticsearch uses a quorum-based master election
  minimum_master_nodes = (num_master_eligible_nodes / 2) + 1
  3-node cluster: minimum_master_nodes = 2 → a partition with 1 node cannot elect a master
  ES 7.0+ replaced this with automated cluster bootstrapping — no manual quorum setting needed

Replica promotion on primary failure:
  Replica that is most up-to-date (tracked by sequence numbers) is promoted
  In-flight writes that were not replicated before failure → lost (durability tradeoff)
  Setting wait_for_active_shards=all before returning write ack reduces loss window

---
Elasticsearch vs Database Full-Text Search

Database LIKE query:
  SELECT * FROM products WHERE description LIKE '%running shoe%'
  Requires full table scan — O(N) rows examined
  No relevance scoring, no ranking
  Usable only for small tables (<100K rows) or with strict prefix patterns (LIKE 'running%')
  Never use LIKE '%term%' on large tables in production

Postgres GIN Index (full-text search):
  Built-in full-text search with tsvector / tsquery
  CREATE INDEX ON products USING GIN(to_tsvector('english', description))
  Inverted index under the hood — fast term lookup
  Supports ranking via ts_rank()
  Good for: up to ~10M documents, moderate query rate, simple ranking
  Limitations: no distributed scaling, no real-time analysis pipeline, limited scoring flexibility

When to use Elasticsearch:
  - Document count > 10M with complex relevance requirements
  - Need BM25 tuning, boosting, multi-field scoring
  - Autocomplete completion suggester at scale
  - Aggregations + search combined (faceted search, filter + sort)
  - Log/event analytics (ELK stack)
  - Distributed horizontal scaling beyond a single DB node

When to use Postgres full-text:
  - Dataset fits on a single DB instance
  - Search is a secondary feature, not the core product
  - Team already owns Postgres, wants to avoid operational complexity of a separate cluster
  - Simple keyword matching without complex relevance tuning

---
Trade-offs

Write amplification:
  Every document write triggers: analysis, segment creation, translog write, replication to replicas
  A single write may update inverted index, stored fields, doc values, BKD tree (for numerics)
  Bulk indexing API (_bulk) amortizes overhead — always prefer bulk for large ingestion jobs
  Cost: typically 5–10x write amplification vs raw storage size

Eventual consistency of index:
  Writes are durable (translog) but not immediately searchable (segment refresh)
  Default 1s refresh → stale reads for up to 1 second after write
  After node failure: replica promotion means in-flight writes may not appear
  Elasticsearch is NOT a transactional database — do not use it as a system of record

Storage overhead:
  Index overhead: inverted index, stored source, doc values (for sorting/aggregations) = typically 2–3x raw data size
  n-gram tokenization for autocomplete: can 5–10x the index size
  Mitigation: disable _source if not needed, use doc_values selectively, use index: false on non-searched fields

---
Real-World Examples

Google autocomplete:
  Offline: query log → MapReduce → ranked trie per language per region → pushed to serving nodes
  Online: prefix lookup in trie (in-memory, sub-millisecond) → top 10 completions
  Cache: per-prefix cache with very high hit rate for common prefixes
  Personalization: blend global completions with per-user query history

Twitter search:
  Two indices: real-time index (last 7 days, in-memory, low latency) + archive index (all tweets, on disk)
  Real-time: incoming tweets indexed within seconds; Lucene-based in-memory store
  Query: search both, merge results by recency + engagement score
  Scale: ~500M tweets/day → write path is the primary challenge

E-commerce product search:
  Inverted index on: title (high boost), description (medium), brand, category
  Faceted search: aggregations on category, brand, price range, rating
  Personalization: boost products the user has viewed or purchased
  A/B testing: run multiple ranking models; measure click-through rate and conversion
  Catalog size: Amazon has ~350M products — requires heavy sharding and caching of query results
