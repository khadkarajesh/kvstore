## Reading Dynamo Paper

### End-to-End Flow — How Dynamo Handles a Write

  Client → any node (round robin) → node hashes key → finds coordinator from ring map
  → coordinator replicates to N-1 preference list nodes → waits for W acknowledgements
  → returns success to client → remaining replicas sync asynchronously

  On read:
  Client → any node → coordinator contacts R nodes → if versions match, return value
  → if versions conflict (vector clocks diverge) → return all versions + clocks to client
  → client merges → client writes resolved version back with updated vector clock

  Failure during write:
  Preferred node down → sloppy quorum substitutes next healthy node → hinted handoff
  → original recovers → substitute delivers replica → deletes local copy
  → substitute also crashed → Merkle tree anti-entropy detects and repairs divergence

  This is the narrative to deliver when asked: "Design a highly available key-value store."

---

### Why Dynamo Exists — The Business Motivation

  Amazon's scale means failures are the norm, not the exception. A system that becomes unavailable during node or network failures is unacceptable for the shopping cart — a customer must always be able to add items, even if some nodes are down.

  Core decision: sacrifice consistency, never sacrifice availability.
  - A temporarily stale cart is a minor inconvenience
  - A cart that errors out loses the sale

  This single business requirement drives every design decision in Dynamo: sloppy quorum, hinted handoff, eventual consistency, client-side conflict resolution — all exist to preserve write availability at all costs.

  Interview signal: always anchor Dynamo's mechanisms to this motivation. "Dynamo chose AP because..." is stronger than jumping straight into how consistent hashing works.

---

### Request Routing — Any Node Can Serve Any Request

  Clients can talk to any node in the cluster — no special routing logic required on the client side.

  When a node receives a request:
  - It hashes the key → finds the coordinator from its local ring map
  - If it is the coordinator → handles the request directly
  - If it is not → forwards the request to the coordinator

  Every node has a full copy of the ring map (kept consistent via gossip), so any node can compute the correct coordinator independently.

  Interview signal: "How does a client know which node to talk to?"
  - Client picks any node (round robin or random)
  - That node handles routing internally
  - No central router, no single point of failure in request routing

---

### Paritioning
  Goal: Even key distribution + elasticity (scale nodes in/out without full reshuffling)
  ---
  Plain Consistent Hashing (Physical Nodes Only) — The Problem

  - Each physical node is placed on the ring by hashing its ID → position is random
  - Random positions = uneven arc sizes = some nodes own large keyspace slices
  - Large arc = hotspot (one node handles disproportionate traffic/load)
  - Node removal/addition moves large chunks of keys to a single successor

  ---
  Virtual Nodes — The Fix

  - Each physical node is split into multiple vnodes, scattered across the ring
  - More vnodes per node = smaller, evenly distributed arcs
  - The ring now has many small arcs instead of a few large ones

  On physical node failure:
  All its vnodes are removed. Each vnode's keys move to the next vnode in the ring — which belongs to a different physical node. Because vnodes are scattered, load
   spreads evenly across many physical nodes rather than dumping onto one.

  On add/remove:
  Only the keys in each small vnode arc move — not a large contiguous chunk. Total key movement is minimal.

  ---
  Heterogeneous Hardware

  Higher-capacity machines get more vnodes assigned → own proportionally more keyspace. Lower-capacity machines get fewer. Hardware capability directly maps to
  load share.

  ---
  Key Movements
 
   When a node is added or removed, only K/N keys move — that's 1/N of all keys. The rest stay put. At any cluster size, only one node's fair share of data moves, not a full reshuffle.

  ---
  Key Takeaway

  Vnodes solve three things: uneven distribution, disproportionate failure recovery, and hardware heterogeneity — all through the same mechanism: many small
  scattered arcs instead of few large ones.


  ### Replication

  **§4.3 — How data gets replicated**

  - Coordinator node stores the key + replicates to N-1 clockwise successors
  - That list of N nodes = preference list
  - Preference list holds more than N nodes to account for failures
  - Skips vnodes of the same physical node to ensure N distinct physical nodes

  ---

  **§4.4 — Data Versioning**

  Why it exists: Dynamo is always writable + replicates asynchronously → two nodes can accept writes to the same key simultaneously → divergent versions.

  Important distinction: Dynamo uses vector clocks for conflict DETECTION only — not to achieve causal consistency. A client can still see writes out of causal order across replicas. Vector clocks tell you that divergence happened, not what order writes should be seen in.

  Vector Clocks
  A list of (node, counter) pairs attached to every object version. Each write increments the writing node's counter.

    N1 writes:          [N1:1]
    N1 writes again:    [N1:2]
    N2 writes after:    [N1:2, N2:1]

  Conflict detection rule:
  - Every counter in A ≤ B → B descends from A → no conflict, B wins
  - Any counter in A > B AND any counter in B > A → concurrent writes → conflict

    A: [N1:2, N2:1]
    B: [N1:1, N2:2]
    → Neither dominates → conflict

  ---
  Conflict Resolution

  Dynamo cannot resolve conflicts itself — it doesn't understand your data semantics.

  On get(), all conflicting versions + their vector clocks are returned to the client. Client resolves based on use case:

  - Union: data loss not tolerable (shopping cart — merge both)
  - Last Write Wins: data loss acceptable (use timestamp as tiebreaker)

  Client writes the resolved version back via put(key, context, object) — the context carries the new vector clock establishing causal history.

  ---
  Key Tradeoffs

  Who resolves conflicts — system or client?
  - System (LWW) = simple but silently loses data. 
  - Client (semantic merge) = no data loss but every client must implement merge logic.

  When are conflicts detected — write or read?
  - Dynamo detects on read → writes never block → higher availability. 
  - Some systems detect on write → reads always clean but writes can fail.

  What consistency are you promising?
  - Vector clocks tell you that divergence happened. They don't prevent stale reads. If your use case can't tolerate stale reads, eventual consistency is the wrong model.

  ---
  The Core Insight for Interviews

  Whether eventual consistency is viable depends on your data type:
  - Shopping cart → union merge works → eventual consistency fine
  - Bank balance → union merge loses money → needs stronger consistency

  Always ask: what does merge mean for this data? That answer determines your consistency model.

### Failures

  **Temporary Failures — Sloppy Quorum + Hinted Handoff**

  Preference list = the N nodes designated to store a given key under normal conditions.

  N, W, R Parameters:
  - N = total replicas per key
  - W = minimum nodes that must acknowledge a write
  - R = minimum nodes that must respond to a read
  - W + R > N = overlap guarantees at least one node in read quorum has the latest write
  - Dynamo default: N=3, W=2, R=2

  Tuning tradeoff:
  - Lower W → faster writes, higher availability, more stale read risk
  - Higher R → stronger read consistency, slower reads

  ---

  Sloppy quorum: when some preferred nodes are down, Dynamo substitutes the next healthy nodes clockwise beyond the preference list.

    Normal (N=3):  write to A, B, C
    B goes down:   write to A, C, D  ← D is outside preference list

  Still requires W acknowledgements — but those W nodes don't have to be from the preference list. That's what makes it "sloppy" vs strict quorum.

  Hinted Handoff: substitute node (D) stores the replica with a hint — the name of the original node (B). D periodically attempts to deliver the replica to B once B recovers, then deletes its local copy.

  Multi data-center failure: preference list is constructed to include vnodes from different data centers so a full DC outage doesn't take down all replicas.

  ---

  **Permanent Failures — Merkle Trees (Anti-Entropy)**

  Problem: hinted handoff can fail if the substitute node also crashes before delivering. Replicas diverge with no mechanism to repair.

  Merkle tree: each node maintains a hash tree over its key ranges. Leaf nodes = hashed values. Parent nodes = hash of children. If any value changes, hashes change up the tree.

  Synchronization: two nodes exchange subtree roots to find diverged ranges. Only mismatched subtrees are transferred — not all data. Minimizes data transmitted over the network.

  Tradeoff: frequent writes cause frequent hash recomputation. Merkle trees run in the background (anti-entropy process) — not on the hot path — so recomputation doesn't block reads/writes.

  ---

  **The Failure Chain for Interviews**

  Temporary failure → sloppy quorum handles the write → hinted handoff repairs when node recovers
  Substitute also fails → hinted handoff breaks → Merkle trees detect and repair divergence

  Each mechanism is the safety net for the one before it.

### Membership & Failure Detection (§4.8 — Gossip)

  Every node maintains a local view of the ring. Gossip keeps it accurate without a central coordinator.

  How gossip works:
  - Every second, each node picks a random peer and exchanges membership info
  - They reconcile — whoever has newer info wins
  - Propagates like a rumor: A tells B, B tells C → all nodes converge in O(log N) rounds

  Failure detection:
  - Each node tracks a heartbeat counter for every other node
  - Counter not incremented past threshold → marked temporarily unavailable
  - Prolonged absence → marked permanent failure → Merkle tree anti-entropy kicks in
  - Not instant — transient network issues cause false positives, so detection is intentionally delayed

  Seed nodes:
  - Bootstrap problem: new node joins, nobody knows about it
  - Seed nodes are well-known nodes (hardcoded/config) that all nodes gossip with on startup

  Interview signal: "How does your system detect node failures?"
  - Gossip = decentralized, no single point of failure in detection itself
  - Tradeoff: eventual detection, not instant — node can be down seconds before others notice
  - vs Raft: leader sends heartbeats, follower detects via election timeout — centralized through leader

---

### Production Lessons (§6 — Experiences)

  Real N/W/R tuning at Amazon:
  - Default: N=3, W=2, R=2
  - W=1: maximum write availability, accepted stale reads (non-critical data)
  - W=N, R=1: write to all, read from one — strong durability, fast reads (read-heavy services)

  Conflict rates in production:
  - Divergent versions requiring client resolution were < 1% of reads for shopping cart
  - Built vector clock complexity for a worst case that rarely triggered — but handled it correctly when it did

  Latency numbers:
  - Median read latency: ~5ms
  - 99.9th percentile: ~300ms
  - The gap is the real story — tail latency kills user experience at scale, not the average

  Key insight for interviews:
  - Averages lie. Optimize for tail latency (p99, p999), not median
  - At Amazon's scale, p99 affecting 1 in 1000 users = millions of bad experiences per day
