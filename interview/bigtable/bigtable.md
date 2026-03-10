  Bigtable Reading Guide

  Paper: Bigtable: A Distributed Storage System for Structured Data (Google, 2006)
  URL: https://static.googleusercontent.com/media/research.google.com/en//archive/bigtable-osdi06.pdf

  ---
  What Bigtable Is — Before You Read

  Bigtable is Google's internal large-scale structured storage system. It sits on the CP side of CAP — the direct contrast to Dynamo. Where Dynamo sacrifices
  consistency for availability, Bigtable sacrifices availability for strong consistency and structured data.

  It's the foundation for: HBase, Cassandra's data model, DynamoDB's structured types, Google Cloud Bigtable.

  ---
  Reading Order

  §1 — Introduction
  # Why Google built it, what problems it solves?

  - GFS stored raw files but no structure
  - SQL had structure but vertical scaling only
  - Bigtable fills the gap: structured data + horizontal scale across
    thousands of commodity machines

  - One system, two workloads:
    - Latency-sensitive: Gmail (<100ms per user request)
    - Throughput-sensitive: Web crawl (MapReduce batch jobs)
    - 60+ Google products, no custom system per product

  - Range scan access pattern:
    - Gmail: all emails for user42 this week → scan user42#2024-03-01 to user42#2024-03-07
    - Web crawl: all pages under cnn.com → scan com.cnn.*
    - Time-series: all CPU readings for server42 last hour → scan server42#14:00 to server42#15:00

    - Bigtable: global keyspace range scan
      - Entire table is one globally sorted keyspace
      - Scan can cross any row key boundary freely
      - scan user42#* to user50#* crosses multiple users — one operation

    - Cassandra: range scan within partition only
      - Data is partitioned by partition key (e.g. user_id)
      - Range scan works only on clustering columns within one partition
      - "All emails for user42 sorted by date" → valid, within one partition
      - "All emails across user42 and user43" → requires two separate queries
      - Cross-partition scan = multiple network hops, not supported natively

    - Cassandra: cross-partition range scan (e.g. user12 to user42) requires the developer to query each
  partition separately and merge results. Bigtable: the client library handles cross-tablet scanning
  transparently — developer issues one scan, complexity is abstracted away.

  - Can tolerate brief unavailability in exchange for consistency (CP):
    - Tablet server dies → master reassigns tablets, brief downtime
    - No stale data served — better to show "loading" than show deleted emails

  - What "structured data at Google scale" means?
    - Sparse, Semi-structured rows at petabyte scale, automatic horizontal scaling, no manual sharding
    - GFS doesn't have structure, SQL couldn't autoscale.
    - Sparse, holds multiple version of data, space is not allocated if column doesn't have value, unlike relational database

  §2 — Data Model ← Most important section
   Data Model
    - Components: row key, column family, column, timestamp → cell value
    - Rows stored in global sorted lexicographic order by row key
      → this is what makes range scans fast

    - Column families: declared at table creation, small in number,
      stored together on disk (locality groups)
    - Columns: created on the fly, unlimited, sparse per row
      (missing columns take zero space)

    - Timestamps: assigned by Bigtable (microseconds) or application
    - Multiple versions per cell maintained automatically
    - Garbage collection configured per column family:
      → keep last N versions
      → keep data within last N days

    - Row key design controls data locality:
      → related rows stored physically adjacent
      → enables fast range scans in one sequential disk read
      → bad row key = random distribution = slow scans

  §4 — Building Blocks

  Write Path
  - Write arrives → committed to WAL on GFS first (crash recovery)
  - Written to MemTable (in-memory, mutable, sorted)
  - MemTable full → flushed to SSTable on GFS (immutable, sorted)
  - Separate SSTable per column family
  - MemTable write happens regardless of whether key exists or not

  Read Path
  - Check MemTable first
  - If not found → scan SSTables newest to oldest
  - Bloom filter used per SSTable to skip files that don't contain the key
  - Avoids sequential scan of every SSTable

  Compaction
  - Minor:   MemTable flushed to new SSTable
  - Merging: few SSTables merged into one
  - Major:   all SSTables merged, deleted data physically removed

  Three Systems — One Job Each
  - Chubby:        who is the master?
  - Master:        which tablet server owns which tablet?
  - Tablet Server: read and write on assigned tablets

  Chubby (distributed lock service)
  - Stores root tablet location
  - Ensures only one master is active at a time
  - Maintains session per tablet server
    → session timeout = master considers tablet server dead
    → master reassigns that server's tablets
  - When master fails: waiting process acquires Chubby lock
    and promotes itself — Chubby does not assign, it only holds the lock

  Master Server
  - Assigns tablets to tablet servers
  - Detects dead tablet servers via Chubby session expiry
  - Handles load balancing
  - NOT on the read/write path — client never talks to master

  Client Lookup — Three Level Hierarchy
  - Client → Chubby (where is root tablet?)
           → Tablet Server with root tablet (where is METADATA tablet?)
           → Tablet Server with METADATA tablet (where is my data?)
           → Tablet Server with actual data directly
  - Client caches tablet locations — steps 1-3 skipped on subsequent reads
  - Master is never in this flow

  GFS
  - Stores all SSTables and WAL files
  - Tablet servers do not own their data — GFS does
  - When tablet server dies, master reassigns tablets immediately
    because data is already safely on GFS

  

  §5 — Implementation

  How Read Happens
  - Client checks cache:
    - Tablet address exists → request directly to tablet server
      - Row key checked in MemTable first → if found, return to client
      - If not found → scan SSTables newest to oldest
      - Bloom filter used per SSTable to skip files that don't contain the key
    - Tablet address missing or invalidated:
      - Client → Chubby (root tablet location)
      - Client → root tablet server (METADATA tablet location)
      - Client → METADATA tablet server (user tablet server address)
      - Client → user tablet server directly, address cached for future use
  - Prefetching optimization: client fetches row keys above and below
    the target key to reduce future network round trips
  - Cache invalidated when:
    - Tablet server fails
    - Master fails
    - Tablet splits

  How Write Happens
  - Authorization check: tablet server verifies client has write
    permission via Chubby ACL file
  - Write logged to WAL on GFS first (durability, crash recovery)
  - Then written to MemTable (in-memory)
  - MemTable full → minor compaction: MemTable frozen, new MemTable
    created for incoming writes, old MemTable flushed to new SSTable on GFS

  Compactions
  - Minor:   MemTable → one new SSTable
             does NOT remove deleted records
  - Merging: few SSTables + current MemTable → one new SSTable
             does NOT remove deleted records
  - Major:   ALL SSTables merged → one SSTable
             deleted records physically removed ← only major does this

  Master Server — what it does and does NOT do
  - Assigns tablets to tablet servers
  - Detects dead tablet servers via Chubby session expiry:
    → tablet server maintains Chubby session keepalive
    → session expires → tablet server kills itself (avoids split-brain)
    → master detects dropped session, deletes server from directory,
      reassigns its tablets — master does NOT acquire the tablet server's lock
  - Tablet splitting:
    → tablet server decides when tablet is too large
    → tablet server splits it, records new tablet info in METADATA
    → tablet server notifies master
    → master assigns the new tablet
  - NOT on the read/write path — client never talks to master

  Master Startup / Recovery
  - Master does NOT store commits or use a WAL
  - On startup, master rebuilds state by:
    1. Acquiring master lock in Chubby (ensures only one master)
    2. Scanning Chubby server directory to find live tablet servers
    3. Asking each tablet server which tablets it currently serves
    4. Scanning METADATA tablets to find unassigned tablets
  Master Election (how new master is assigned)
  - All nodes in the cluster run a master process (standbys)
  - On startup all processes try to acquire /master-lock in Chubby
    → only one succeeds → becomes master
    → others register a Chubby watch on /master-lock ("notify me when free")
    → they block, no busy polling
  - Current master keeps lock alive via periodic heartbeats to Chubby
    → Master → Chubby: heartbeat every N seconds (I am alive)
    → Chubby → Master: acknowledged
  - When master crashes:
    → heartbeats stop
    → Chubby waits for lease timeout (~10s)
    → session expires → /master-lock released automatically
    → Chubby notifies waiting processes
    → one acquires the lock → becomes new master
  - Lease timeout is intentional:
    → too short = unnecessary failovers on slow network
    → too long = unnecessary downtime on real failure

  Chubby Namespace (what Bigtable stores in Chubby)
  - /master-lock        → master election lock
  - /root-tablet        → root tablet location
  - /servers/           → server directory
      tablet-server-1   → session file (exists = server alive)
      tablet-server-2   → session file
      tablet-server-3   → session file
  Server directory is part of Chubby — not a separate system

  Tablet Assignment Flow (normal path)
  - Master sends load request to a tablet server
  - Tablet server accepts → reads tablet info from METADATA tablet
  - Loads SSTables from GFS
  - Replays WAL to rebuild MemTable
  - Starts serving reads and writes for that tablet

  Load Balancing
  - Master periodically checks if tablet servers are over/under loaded
  - To move a tablet:
    1. Master sends unload request to source tablet server
    2. Source server stops serving, flushes MemTable to GFS
    3. Master sends load request to destination tablet server
    4. Destination loads SSTables + replays WAL → starts serving

  Tablet Server Recovery (when tablet server fails)
  - SSTables are already safely persisted on GFS — no recovery needed
  - MemTable is reconstructed by replaying the WAL (commit log on GFS)
    → WAL contains writes that were in memory but not yet flushed to SSTable
  - New tablet server reads METADATA to find its assigned tablets,
    loads SSTables from GFS, replays WAL to rebuild MemTable

  §6 — Refinements
  - Locality groups — performance optimization
    - Locality groups on column family helps to optimize the performance
    - For example content of the email might not be required to fetch at the time of fetching metadata of gmail.
    - Better to group with locality of data
  - Bloom filters — read optimization
   - Bloom filters helps to optimize the read by checking the row key existance in the data structure present in the memory, it doesn't have to do file scan
  - Caching — two levels
    Two levels of Cache in Tablet server to optimize the read
    - Scan Cache - cache of the key value pairs read from SSTables, best for frequent key value read usage type application
    - Block Cache - Cache of the SSTables block, best for the application which requires near by data as well, for analytics application.
    - 

  ---
  Sections to Skip

  - §3 (API) — skim, 1 minute
  - §7 (Performance numbers) — skim the summary only
  - §8 (Real applications) — skip

  ---
  Key Questions to Answer While Reading

  1. What are rows, column families, columns, and timestamps — and how do they relate?
  2. What is an SSTable and how does a read find data across multiple SSTables?
  3. What does the master server do — and what does it NOT handle?
  4. What is a tablet and how does Bigtable scale horizontally?
  5. What is Chubby and why does Bigtable need it?

  ---
  Retention Questions When You Return

  1. A row in Bigtable can have thousands of columns, and two rows don't need the same columns. How is this possible and what problem does it solve?
  2. Bigtable is CP, Dynamo is AP. What does that mean concretely for how each handles a node failure?
  3. The master server is not on the read/write path. How does a client find the right tablet server then?