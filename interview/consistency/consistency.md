## Consistency Models

---

### The Core Question
After a write happens — who is guaranteed to see it, and when?

---

### 1. Eventual Consistency

**Guarantee:** All replicas converge eventually. No timing promise. Stale reads are possible.

**Real World:**
- Amazon shopping cart — stale cart briefly is fine, availability is more important
- DNS propagation — domain update takes minutes to propagate globally
- Social media like/view counts — showing 10,002 vs 10,003 doesn't matter
- Product catalog — price showing old value for a few seconds is acceptable

**How it's achieved:**
- Writes are acknowledged after W nodes respond — no waiting for all replicas
- Replication happens asynchronously in the background
- Anti-entropy (Merkle trees) periodically syncs diverged replicas
- Gossip protocol propagates membership and state changes across nodes
- No coordination on reads — any replica responds immediately

**Interview signal:** "High availability", "can tolerate stale reads", "always writable" → eventual consistency.

---

### 2. Session Consistency (Read Your Own Writes)

**Guarantee:** You always see your own writes immediately. Other clients may still see stale data.

**Real World:**
- Twitter — you see your own tweet immediately after posting
- User profile update — you change your bio, you see it reflected right away
- E-commerce order — you place an order, your order history shows it immediately
- Reddit — you see your own comment appear right after submitting

**How it's achieved:**

**Option 1 — Sticky Sessions via Consistent Hashing on User ID**
- Hash the user ID to a position on the ring → always routes to the same node
- Same mechanism as data partitioning, applied to request routing
- Problem: if that node goes down, user loses the guarantee until session is re-established

**Option 2 — Session ID in Cookie (Load Balancer Level)**
- Client gets a session ID on first request
- Load balancer reads session ID from cookie → routes to the same backend node
- Simpler than consistent hashing but doesn't distribute load as evenly
- Problem: node failure breaks the session affinity

**Option 3 — Version Tokens (Most Flexible)**
- After a write, server returns a version token (a timestamp or log sequence number)
- Client attaches this token to subsequent reads
- Any node can serve the read as long as it has caught up to that version
- Node checks: "do I have version ≥ token?" → if yes, serve; if no, wait or redirect
- Advantage: not tied to a specific node, survives node failures
- Used by: DynamoDB's strongly consistent reads, Google Spanner

**Interview signal:** "User must see their own changes immediately" → session consistency.

---

### 3. Causal Consistency

**Guarantee:** If write B happened because of write A (causally related), everyone sees A before B. Unrelated writes can be reordered freely.

**Real World:**
- Facebook comments — reply must appear after the original post for everyone
- Facebook Live chat — messages in a thread must appear in causal order
- Google Docs — your edit responding to someone else's must appear after theirs
- WhatsApp threads — reply ordering within a conversation must be preserved
- Collaborative editors — if you delete a line and I edit it, I must see deletion first

**How it's achieved:**

**Vector Clocks (Dynamo's approach)**
- Every write carries a vector clock: list of (node, counter) pairs
- Client sends its current vector clock with every request
- Node only serves a read after it has applied all writes the client has seen
- If node is behind, it waits or redirects to a node that is caught up
- Tracks causal dependency: "I saw [N1:2, N2:1] before writing this"

**Causal+ Consistency (COPS system)**
- Extended vector clocks with explicit dependency tracking
- Write to local datacenter first → replicate with dependency metadata
- Remote datacenter applies write only after all its dependencies are satisfied
- Used in geo-distributed systems where global ordering is too expensive

**Interview signal:** "Operations are related/dependent", "ordering within a context matters" → causal consistency.

---

### 4. Strong Consistency (Linearizability)

**Guarantee:** Every read reflects the most recent write. System behaves as if there is one copy of data. Operations appear instantaneous in a single global order.

**Real World:**
- Bank transfers — debit and credit must be atomic and immediately visible to everyone
- Flight/concert seat booking — two users cannot book the same seat
- Inventory management — cannot oversell the last item in stock
- Distributed locks (Zookeeper) — only one node can hold the lock at a time
- Leader election (Raft) — only one leader per term
- Financial ledgers — every transaction reflects true current balance

**How it's achieved:**

**Quorum Reads + Writes (Strict, not Sloppy)**
- W + R > N ensures overlap — at least one node in read quorum has latest write
- Unlike Dynamo's sloppy quorum, strict quorum only uses preference list nodes
- If quorum unreachable → operation blocks or fails (no substitutes allowed)
- Problem: partition means unavailability

**Raft / Paxos Consensus**
- Every write goes through a single leader
- Leader replicates to majority before acknowledging write to client
- Reads also go through leader (or nodes verify they are still leader)
- No stale reads possible — leader always has latest committed value
- Partition: minority partition rejects all writes, majority partition continues
- Your KV store uses this

**Single Leader with Synchronous Replication**
- Leader waits for all replicas to acknowledge before returning success
- Stronger than Raft quorum — every node is up to date, not just majority
- Problem: one slow replica blocks all writes — high latency, low availability

**Two-Phase Commit (Distributed Transactions)**
- Coordinator asks all nodes to prepare → all nodes lock and confirm ready
- Coordinator commits → all nodes apply
- Used when write spans multiple nodes/shards (bank transfer across accounts)
- Problem: coordinator failure leaves nodes locked — availability risk

**Interview signal:** "Cannot tolerate data loss", "exactly once", "inventory", "locks", "financial" → strong consistency.

---

### Decision Flowchart

```
Can the business tolerate stale reads?
├── No → Strong consistency (Raft, Zookeeper, Spanner)
└── Yes → Does operation ordering matter between related writes?
           ├── Yes → Causal consistency (vector clocks, COPS)
           └── No → Does the user need to see their own writes immediately?
                    ├── Yes → Session consistency (sticky sessions, version tokens)
                    └── No → Eventual consistency (Dynamo, Cassandra)
```

---

### Consistency vs Availability Tradeoff (CAP)

Network partitions always happen at scale. Real choice is C vs A:

| Model | CAP choice | Partition behavior |
|---|---|---|
| Eventual | AP | Stays available, accepts stale reads |
| Session | AP (weak) | Stays available, guarantee may break on failover |
| Causal | AP (stronger) | Stays available, delays causally dependent reads |
| Strong | CP | Blocks or rejects operations until partition heals |

---

### Consistency Spectrum

```
Eventual → Session → Causal → Strong (Linearizable)
  ↑                                      ↑
Highest availability              Highest correctness
Lowest coordination               Highest coordination
Lowest latency                    Highest latency
```

Each step up = more coordination between nodes = more latency = less availability.
