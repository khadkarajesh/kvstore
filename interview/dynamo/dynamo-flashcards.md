## Dynamo Paper — Anki Flashcards

Format: Q → A
Cover the answer, recall it, then check.

---

**Q: Why did Amazon choose AP over CP for Dynamo?**
A: Failures are the norm at scale. A shopping cart must always accept writes — a stale cart is a minor inconvenience, an error loses the sale.

---

**Q: What is a preference list?**
A: The list of N nodes designated to store a given key, determined by walking clockwise on the ring from the key's hash position. Contains more than N nodes to account for failures.

---

**Q: What is sloppy quorum?**
A: When preferred nodes are down, Dynamo substitutes the next healthy nodes clockwise beyond the preference list. Still requires W acknowledgements — just not from the "correct" nodes.

---

**Q: What is hinted handoff?**
A: The substitute node stores the replica with a hint (name of original node). When the original recovers, the substitute delivers the replica and deletes its local copy.

---

**Q: What is a vector clock?**
A: A list of (node, counter) pairs attached to every object version. Each write increments the writing node's counter. Tracks the causal history of writes.

---

**Q: When do two vector clocks indicate a conflict?**
A: When any counter in A > B AND any counter in B > A — neither clock dominates the other. Both represent concurrent writes.

---

**Q: When does one vector clock dominate another (no conflict)?**
A: When every counter in A ≤ every counter in B — B descends from A, B wins, no conflict.

---

**Q: Who resolves conflicts in Dynamo?**
A: The client. Dynamo returns all conflicting versions + their vector clocks. Client merges based on application semantics (union for cart, LWW for non-critical data).

---

**Q: What is Dynamo's default N, W, R?**
A: N=3, W=2, R=2

---

**Q: What does W + R > N guarantee?**
A: At least one node in the read quorum participated in the write quorum — overlap ensures the latest write is always reachable on a read.

---

**Q: What does lowering W do?**
A: Faster writes, higher write availability, more stale read risk.

---

**Q: What does raising R do?**
A: Stronger read consistency (larger overlap with write quorum), slower reads.

---

**Q: N=3, W=2, R=2. One node down — can you read? Write?**
A: Yes to both. W=2 and R=2 can still be satisfied with 2 remaining nodes.

---

**Q: N=3, W=2, R=2. Two nodes down — can you read? Write?**
A: No to both. Only 1 node remains, cannot satisfy W=2 or R=2.

---

**Q: Why do vnodes exist instead of placing physical nodes directly on the ring?**
A: Physical nodes create random, uneven arc sizes → hotspots. Vnodes scatter each physical node across many small arcs → even distribution, proportional hardware allocation, minimal key movement on failure.

---

**Q: When a physical node fails, where do its keys go?**
A: Each vnode's keys move to the next vnode clockwise on the ring — which belongs to a different physical node. Load spreads evenly across many nodes, not dumped on one successor.

---

**Q: How much data moves when a node joins or leaves?**
A: K/N keys — that's 1/N of all keys. Only the departing node's fair share moves. Everything else stays put.

---

**Q: How does gossip work in Dynamo?**
A: Every second, each node picks a random peer and exchanges membership info. They reconcile — newer info wins. All nodes converge in O(log N) rounds.

---

**Q: How does Dynamo detect node failure?**
A: Each node tracks heartbeat counters for all others. Counter not updated past a threshold → temporarily unavailable. Prolonged absence → permanent failure → Merkle tree anti-entropy starts.

---

**Q: What are seed nodes?**
A: Well-known nodes (hardcoded or via config) that all nodes gossip with on startup. Solves the bootstrap problem when a new node joins and nobody knows about it.

---

**Q: Why are Merkle trees used in Dynamo?**
A: To detect and repair diverged replicas efficiently. Two nodes exchange subtree hashes — only mismatched subtrees are transferred, not all data. O(differences), not O(data size).

---

**Q: What happens if hinted handoff fails (substitute also crashes)?**
A: Merkle tree anti-entropy (background process) detects the divergence and repairs it.

---

**Q: What is the failure recovery chain in Dynamo?**
A: Temporary failure → sloppy quorum handles write → hinted handoff repairs on recovery → substitute crashes → Merkle trees detect and repair divergence.

---

**Q: How does any node know which node is the coordinator for a key?**
A: Every node has a full copy of the ring map (kept consistent via gossip). Any node hashes the key and walks clockwise to find the coordinator independently.

---

**Q: What was Dynamo's conflict rate in production for the shopping cart?**
A: < 1% of reads returned multiple conflicting versions. Built for the worst case that rarely triggered.

---

**Q: Dynamo median read latency vs p99?**
A: Median ~5ms, p99 ~300ms. The gap is what matters — tail latency determines user experience at scale, not the average.

---

**Q: Why does Dynamo detect conflicts on read, not on write?**
A: Detecting on read means writes never block → higher write availability. Detecting on write would require coordination at write time → slower writes, lower availability.

---

**Q: What is the context parameter in put(key, context, object)?**
A: The vector clock of the version being updated. Dynamo uses it to place the new write correctly in the causal history. Without it, Dynamo can't tell if the write is an update or a concurrent branch.

---

## Consistency Models

---

**Q: What is the core question that separates consistency models?**
A: After a write happens — who is guaranteed to see it, and when?

---

**Q: What does eventual consistency guarantee?**
A: All replicas converge eventually. No timing promise. Stale reads are possible for anyone.

---

**Q: What does session consistency (read your own writes) guarantee?**
A: You always see your own writes immediately. Other clients may still see stale data.

---

**Q: What does causal consistency guarantee?**
A: If write B happened because of write A, everyone sees A before B. Unrelated writes can be reordered freely.

---

**Q: What does strong consistency (linearizability) guarantee?**
A: Every read reflects the most recent write. System behaves as if there is one copy of data.

---

**Q: Three ways to implement read your own writes — name them.**
A: 1) Sticky sessions via consistent hashing on user ID. 2) Session ID in cookie routed by load balancer. 3) Version tokens — client attaches write version to reads, any node that has caught up can serve.

---

**Q: What is the problem with sticky sessions when a node goes down?**
A: The user loses the read-your-own-writes guarantee until their session is re-established on a new node.

---

**Q: Why are version tokens better than sticky sessions?**
A: Not tied to a specific node — any node that has caught up to the version can serve the read. Survives node failures.

---

**Q: Does Dynamo use vector clocks to achieve causal consistency?**
A: No. Dynamo uses vector clocks for conflict detection only. It does not guarantee causal consistency — a client can still see writes out of causal order across replicas.

---

**Q: What is the CAP choice for each consistency model?**
A: Eventual = AP. Session = AP (weak). Causal = AP (stronger). Strong = CP.

---

**Q: What happens to each consistency model during a network partition?**
A: Eventual/Session/Causal — stay available, accept some staleness. Strong — blocks or rejects operations until partition heals.

---

**Q: When would you use eventual consistency over strong consistency?**
A: When the data type supports semantic merge (shopping cart, counters, social likes) and stale reads are acceptable. Business cannot tolerate write unavailability.

---

**Q: When is eventual consistency the wrong model?**
A: When stale reads cause correctness violations — bank balance, inventory count, seat booking. Use strong consistency instead.

---

**Q: How is causal consistency implemented?**
A: Vector clocks — client sends its current vector clock with every request. Node only serves the read after applying all writes the client has seen.

---

**Q: How is strong consistency achieved in Raft?**
A: Every write goes through the leader, committed to majority before acknowledging. Reads go through leader. No stale reads possible.

---

**Q: What is two-phase commit and when is it used?**
A: Coordinator asks all nodes to prepare → all lock and confirm → coordinator commits. Used when a write spans multiple nodes/shards (e.g. bank transfer across accounts).

---

**Q: What is the risk of two-phase commit?**
A: Coordinator failure leaves all nodes locked — availability risk until coordinator recovers.
