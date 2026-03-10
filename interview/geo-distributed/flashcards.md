## Geo-Distributed Systems — Flashcards

Format: Q → A
Cover the answer, recall it, then check.

---

**Q: Active-active vs active-passive — what is the core difference?**
A: Active-passive: one region handles all traffic; standby is idle and takes over on failure. Active-active: multiple regions each handle reads and writes simultaneously, providing lowest latency for global users. Active-passive is simpler with no conflict risk; active-active maximizes utilization and availability but requires conflict resolution.

---

**Q: Why is active-active inherently harder than active-passive?**
A: Concurrent writes to the same data in different regions create conflicts — two regions both accepted a write, now replication brings both to each region. You must decide which version wins (or merge them). Active-passive has a single authoritative region, so conflicts cannot occur.

---

**Q: You are building a global e-commerce platform. A product's inventory count is updated by warehouse systems in the US and EU simultaneously. What conflict resolution strategy do you use?**
A: LWW (last-write-wins) is risky here — losing a decrement could oversell inventory. Use application-level conflict resolution with version vectors to detect conflicts, then apply domain logic (sum the deltas, not pick a winner). Or home inventory to a single region (write localization) and accept the latency trade-off.

---

**Q: What is RPO and RTO? Give concrete definitions.**
A: RPO (Recovery Point Objective): how much data loss is acceptable — the maximum time window of data that can be lost on failure. RPO=0 means zero data loss. RTO (Recovery Time Objective): how long the system can be unavailable during recovery. RTO=0 means instant failover with no downtime.

---

**Q: Synchronous vs asynchronous cross-region replication — what is the RPO difference?**
A: Synchronous: write acknowledged only after committed in all target regions → RPO=0 (zero data loss), but write latency includes cross-region RTT (~100ms+). Asynchronous: write acknowledged locally, replication happens later → lower write latency but RPO > 0 (data written since last replication can be lost if primary fails).

---

**Q: Why is synchronous cross-region replication impractical for most user-facing writes?**
A: Cross-continent RTT is ~100-200ms. A synchronous write must wait for an ack from each region before responding to the user. This adds 100-200ms to every write — unacceptable for web applications. Most systems use async replication and accept a small RPO.

---

**Q: What is GeoDNS and how does it route a user to the nearest region?**
A: GeoDNS returns different DNS records based on the geographic location of the DNS resolver making the query. The authoritative DNS server maps the resolver's IP to a region and returns the corresponding load balancer IP. The user's subsequent TCP connection goes to the nearest regional cluster.

---

**Q: Why might GeoDNS route a user to the wrong region?**
A: GeoDNS resolves based on the resolver's IP, not the user's IP. A user in Tokyo using a corporate VPN with a resolver in New York would be routed to the US region. Similarly, public resolvers (8.8.8.8) may be geographically associated with a location different from the actual user.

---

**Q: What is DNS TTL's impact on failover speed and how do you manage it?**
A: DNS TTL is how long clients cache the resolved IP. TTL=300s means a failed region change takes up to 5 minutes to reach all clients. Best practice: lower TTL to 60s before a planned failover; keep 300s normally (less DNS load). Note: some clients ignore low TTLs, so even TTL=60s doesn't guarantee instant propagation.

---

**Q: What is Anycast routing and how does it differ from GeoDNS?**
A: Anycast: multiple servers in different locations advertise the same IP via BGP; the internet's routing layer automatically sends packets to the nearest one. No DNS tricks needed — the routing is at the network layer. GeoDNS: same hostname resolves to different IPs per region — operates at the DNS/application layer. Anycast has no DNS TTL delay but cannot do application-aware routing.

---

**Q: What is a CRDT and give one concrete example?**
A: Conflict-free Replicated Data Type — a data structure where concurrent updates from multiple replicas can always be merged deterministically without conflicts. Example: G-Counter (grow-only counter) — each region maintains its own counter; the total is the sum of all regional counters. Merge = take the maximum per-region value. No conflict possible because the structure only grows.

---

**Q: When is last-write-wins (LWW) appropriate and what is its main risk?**
A: LWW is appropriate when occasional loss of concurrent updates is acceptable (social media likes, view counts, leaderboard scores). Main risk: clock skew — servers in different regions have clocks offset by 10-500ms, so a "later" write with a slightly earlier clock time gets incorrectly discarded. Use logical clocks or hybrid logical clocks (HLC) to reduce this risk.

---

**Q: What are vector clocks and what problem do they solve over LWW?**
A: Vector clocks track causality between events across replicas. Each value carries a version vector {regionA: 3, regionB: 2} updated on every write. Two updates can be compared: one causally precedes the other (safe to discard older), or they are concurrent (conflict — must be resolved). LWW can discard a causally later update due to clock skew; vector clocks detect this case.

---

**Q: What is write localization and why does it reduce conflicts in active-active?**
A: Assign each data object a "home region" based on the user or entity that owns it. Writes for that object are authoritative in the home region; other regions replicate reads. Conflicts only arise if the same object is written in two regions simultaneously — write localization makes this rare by routing writes to a consistent primary region per object.

---

**Q: What is data residency and why does it complicate geo-distributed system design?**
A: Data residency requires that certain data (e.g. EU citizen PII under GDPR) be stored and processed only within specific geographic boundaries. In a geo-distributed system, you cannot freely replicate this data to all regions. You must geo-shard by jurisdiction, restrict which regions can store which data, and ensure backups and DR also respect residency constraints.

---

**Q: A user signs up in Germany, then moves to the US and changes their account's home region. What data residency challenges arise?**
A: All data created during the EU period was stored in the EU region (GDPR compliant). Migrating to US region may violate GDPR for that historical data. Options: keep historical EU data in EU, only create new data in US; or perform a formal data export/deletion under GDPR right to erasure before US storage. Must also ensure migration does not create a window where data exists in both regions.

---

**Q: Cross-continent RTT is ~100ms. Your Raft cluster has nodes in New York, London, and Tokyo. What is the minimum write latency?**
A: Raft requires a quorum (majority) of nodes to acknowledge each write. With 3 nodes, 2 must ack. The slowest quorum pair determines latency. NY-London RTT is ~80ms; NY-Tokyo is ~170ms. A write from NY waits for the first quorum response: ~80ms (London responds before Tokyo). Minimum write latency is ~80ms — too slow for user-facing writes, acceptable for backend coordination.

---

**Q: What is automatic failover vs manual failover and when do you prefer each?**
A: Automatic: health checks detect failure → DNS or LB update without human intervention. Faster RTO (seconds to minutes). Risk: false positives can cause unnecessary failovers (split-brain). Manual: human reviews alerts and executes runbook. Safer for ambiguous failures (network blip vs real outage). Slower RTO. Prefer automatic for clear hardware/software failures with reliable health checks; manual for ambiguous situations or when split-brain risk is high.

---

**Q: How does AWS Global Accelerator differ from CloudFront for global routing?**
A: CloudFront is a CDN — it caches content at edge locations and serves cached responses. Global Accelerator does not cache; it uses Anycast to route traffic to the nearest AWS PoP, then forwards it over AWS's private backbone to your origin. Global Accelerator improves latency for dynamic/non-cacheable traffic by avoiding unpredictable public internet routing. CloudFront is for cacheable content; Global Accelerator is for origin-bound traffic.

---

**Q: Give the approximate RTT for: within a data center, cross-AZ, US coast-to-coast, US to Europe.**
A: Within a data center (same AZ): ~0.1-1ms. Between AZs in the same region: ~1-3ms. US coast-to-coast (us-east-1 to us-west-2): ~50-70ms. US to Europe (US to eu-west-1): ~80-120ms. These numbers matter for evaluating whether synchronous replication or consensus across a given topology is feasible.

---
