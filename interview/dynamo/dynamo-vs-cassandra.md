## Dynamo vs Cassandra

Cassandra is directly inspired by Dynamo (partitioning + replication model) and BigTable (data model + storage engine). Understanding the differences is a common interview follow-up.

---

### What Cassandra Took from Dynamo

- Consistent hashing with vnodes for partitioning
- N, W, R tunable replication (called consistency levels in Cassandra)
- Gossip protocol for membership and failure detection
- Sloppy quorum + hinted handoff for availability during failures
- Eventual consistency as the default model

---

### Key Differences

| | Dynamo | Cassandra |
|---|---|---|
| **Data model** | Pure key-value (blob values) | Wide-column (tables, rows, typed columns) |
| **Conflict resolution** | Client-side (semantic merge via vector clocks) | Last Write Wins (timestamp-based, by default) |
| **Vector clocks** | Yes — detects concurrent writes, returns all versions to client | No — dropped in favor of LWW; simpler but risks data loss |
| **Query model** | get(key) / put(key) only | CQL (SQL-like) — supports range queries, secondary indexes |
| **Open source** | No — internal Amazon system | Yes — Apache Cassandra |
| **Managed service** | Amazon DynamoDB (inspired by, but different) | AWS Keyspaces, Datastax Astra |
| **Coordinator** | Any node can coordinate | Any node can coordinate (same model) |

---

### Why Cassandra Dropped Vector Clocks

Cassandra chose LWW over vector clocks because:
- Vector clocks require clients to handle conflict resolution — adds application complexity
- LWW is simpler operationally — system resolves conflicts automatically using timestamps
- Tradeoff: silent data loss when concurrent writes happen — the "losing" write disappears

For most use cases Cassandra targets (time-series, logs, event data), LWW is acceptable because writes are rarely truly concurrent on the same key.

---

### Why This Matters for Interviews

If asked "would you use Dynamo or Cassandra?":
- **Dynamo-style** (or DynamoDB): when you cannot tolerate data loss and your application can implement merge logic
- **Cassandra**: when you need a richer data model (CQL queries), and LWW data loss is acceptable for your use case

The real question is: **can your data type tolerate silent overwrites?**
- Shopping cart: No → need merge → Dynamo-style
- Sensor readings / time-series: Yes → LWW fine → Cassandra

---

### Amazon DynamoDB vs Dynamo Paper

Common interview confusion — these are different systems.

**The simple summary:**
Dynamo is the internal system described in the 2007 paper, built to solve one problem: shopping cart must never fail to write. DynamoDB is a fully rebuilt managed service inspired by Dynamo's ideas but redesigned to serve thousands of different use cases.

They share DNA — not an implementation.

**What they share:**
- Consistent hashing for partitioning
- Replication across multiple nodes
- High availability as the primary goal
- Key-value storage as the core model

**Where they diverge:**

| | Dynamo (paper) | Amazon DynamoDB (service) |
|---|---|---|
| **Year** | 2007 | 2012 |
| **Conflict resolution** | Client-side vector clocks — app controls merge | LWW by default — simpler, risks silent data loss |
| **Consistency** | Eventual only | Tunable — eventual (default) or strongly consistent reads (2x cost) |
| **Data model** | Pure blob key-value — values are opaque bytes | Structured — typed attributes, nested JSON, lists, sets |
| **Query capability** | get(key) / put(key) only | Primary key + sort key, secondary indexes, range queries |
| **Operations** | Internal only — Amazon engineers operated it | Fully managed — AWS handles replication, scaling, failover, backups |
| **Sloppy quorum** | Explicitly described in paper | Internal implementation not disclosed — likely evolved |

**Why they diverged:**
Dynamo optimized for one use case — always-writable shopping cart. Every decision served that. DynamoDB needed to be a general-purpose database for thousands of teams with different needs — so it added tunable consistency, richer data models, and secondary indexes while keeping the high availability foundation.

**Interview clarification to make:**
> "Are you referring to the 2007 Dynamo paper or Amazon DynamoDB the managed service? They share the same partitioning and replication philosophy but differ significantly in consistency model, data model, and conflict resolution."

That one clarification signals you actually know both systems.
