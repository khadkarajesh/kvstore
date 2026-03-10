## Bigtable Paper — Flashcards

Format: Q → A
Cover the answer, recall it, then check.

---

**Q: What gap does Bigtable fill?**
A: GFS stored raw files but had no structure. SQL had structure but couldn't scale horizontally. Bigtable fills the gap: structured data + automatic horizontal scale across thousands of machines.

---

**Q: What is a tablet?**
A: A contiguous range of rows stored together, managed as one unit, assigned to exactly one tablet server at a time. Tablets split automatically when too large.

---

**Q: What is an SSTable?**
A: A sorted, immutable file of key-value pairs stored on GFS. Once written, never modified — only replaced by compaction. The actual key stored is row key + column family + column + timestamp.

---

**Q: What are the four components of Bigtable's data model?**
A: Row key, column family, column, timestamp → cell value. The intersection of all four is a cell — the atomic unit of storage.

---

**Q: What is the difference between a column family and a column?**
A: Column families are declared at table creation, small in number, stored together on disk. Columns are created on the fly, unlimited in number, sparse per row — missing columns take zero space.

---

**Q: Who assigns timestamps in Bigtable?**
A: Either Bigtable (microseconds since epoch) or the application. Application-assigned timestamps are used when the timestamp has semantic meaning — e.g. web crawl timestamp = when the page was crawled.

---

**Q: How does garbage collection work in Bigtable?**
A: Configured per column family. Two options: keep last N versions, or keep data within last N days. Bigtable handles deletion automatically — no application code needed.

---

**Q: What does single-row atomicity mean in Bigtable?**
A: Reads and writes to a single row are atomic. Multi-row transactions are not supported. This is a deliberate tradeoff — cross-row coordination would require distributed locking and hurt scalability.

---

**Q: Why are rows stored in sorted lexicographic order?**
A: So range scans are fast sequential disk reads. Related rows are physically adjacent. A range scan over a key prefix (e.g. user42#*) hits one contiguous region on disk — no random seeks.

---

**Q: What is the one job of the master server?**
A: Which tablet server owns which tablet. Master assigns tablets, detects dead servers, handles load balancing. It is never on the read/write path.

---

**Q: What is the one job of Chubby in Bigtable?**
A: Who is the master. Chubby also stores root tablet location and maintains tablet server sessions. Normal reads/writes never touch Chubby.

---

**Q: What is the three-level tablet location hierarchy?**
A: Client → Chubby (root tablet location) → root tablet server (METADATA location) → METADATA tablet server (user tablet location) → tablet server with actual data. Client caches all levels — hierarchy skipped on subsequent reads.

---

**Q: Why is the master not on the read/write path?**
A: Scalability. If every read/write went through the master, it would be a bottleneck for billions of requests. Clients find tablet servers directly via the three-level hierarchy and cache the result.

---

**Q: What happens when a tablet server dies?**
A: Tablet server's Chubby session expires → tablet server kills itself (avoids split-brain) → master detects dropped session → master reassigns tablets to other servers → new server loads SSTables from GFS, replays WAL to rebuild MemTable.

---

**Q: What happens when the master dies?**
A: Heartbeats to Chubby stop → Chubby lease timeout (~10s) → /master-lock released → a waiting standby process acquires the lock → promotes itself to master → rebuilds state from Chubby + tablet servers + METADATA.

---

**Q: How does master rebuild its state on startup?**
A: 1) Acquire master lock in Chubby. 2) Scan Chubby server directory to find live tablet servers. 3) Ask each tablet server which tablets it serves. 4) Scan METADATA tablets to find unassigned tablets. Master does NOT use a WAL.

---

**Q: What is the write path in Bigtable?**
A: Authorization check via Chubby ACL → WAL written to GFS (durability) → written to MemTable (in-memory) → MemTable full → minor compaction flushes to new SSTable on GFS.

---

**Q: What are the three compaction types?**
A: Minor: MemTable → one new SSTable (does NOT remove deletes). Merging: few SSTables → one new SSTable (does NOT remove deletes). Major: ALL SSTables → one SSTable (physically removes deleted records).

---

**Q: What is a locality group?**
A: A grouping of column families whose SSTables are stored together on disk. Enables physical separation — reading email metadata never touches body SSTable files. Can be marked in-memory for zero-disk-read access.

---

**Q: What does marking a locality group as in-memory do?**
A: Tablet server loads the entire SSTable for that group into memory permanently. Zero disk reads. Used for METADATA tablet location data which needs to be extremely fast.

---

**Q: What does a bloom filter do in Bigtable?**
A: One bloom filter per SSTable file. Before opening a file, bloom filter answers: "this key is definitely NOT here" → skip the file entirely, no disk I/O. Cannot confirm a key exists — only that it doesn't.

---

**Q: What are the two levels of caching in Bigtable?**
A: Scan Cache: caches key-value pairs — best for repeated reads of same keys. Block Cache: caches raw SSTable blocks — best for sequential/nearby reads (analytics). Scan Cache sits above Block Cache.

---

**Q: Bigtable vs Cassandra — what is the range scan difference?**
A: Bigtable has a global sorted keyspace — scan can cross any row key boundary, client library handles cross-tablet complexity transparently. Cassandra range scans work only within one partition — cross-partition requires multiple queries merged by the application.

---

**Q: When would you pick Bigtable over Cassandra?**
A: When you need range scans across the entire keyspace (not just within a user/partition), petabyte scale, and can tolerate brief unavailability in exchange for consistency (CP).

---

**Q: What is the CP tradeoff in Bigtable?**
A: When a tablet server dies, its tablets are unavailable until master reassigns them. No stale data is served — better to show "loading" than show data that may be inconsistent. Opposite of Dynamo's AP approach.

---

**Q: Why does Bigtable store web crawl row keys as reversed domains?**
A: Rows are sorted lexicographically. Reversed domain (com.cnn.www) groups all pages from the same domain adjacent on disk. A scan of com.cnn.* retrieves all CNN pages in one sequential read. Forward domains would scatter them randomly.

---

**Q: What is the Chubby namespace structure Bigtable uses?**
A: /master-lock (master election), /root-tablet (root tablet location), /servers/ directory (one session file per tablet server — exists = server alive, gone = server dead).
