Geo-Distributed Systems — System Design Notes

---
Active-Active vs Active-Passive

Active-Passive (primary/standby):
  One region handles all traffic; secondary region is a warm (or cold) standby
  Reads and writes go to the primary; standby receives replicated data but serves no traffic
  On primary failure: failover promotes standby to primary

  Trade-offs:
    + Simple — no conflict resolution needed; single authoritative source of truth
    + Well-understood consistency (standby is a replica, not a peer)
    - Users far from primary experience high latency on all requests
    - Underutilizes standby capacity (you're paying for idle resources)
    - Failover requires DNS change or load balancer update → downtime window

  When to use:
    - Writes require strong consistency and conflicts are unacceptable
    - Lower operational complexity is a priority
    - Regulatory: data must not be processed outside one region (but standby can be in another for DR)

Active-Active (multi-primary):
  Multiple regions each handle reads and writes simultaneously
  Users routed to nearest region; data replicated asynchronously between regions

  Trade-offs:
    + Lowest latency — users served from nearest region
    + Full utilization of all regional capacity
    + No single region failure takes down the whole service
    - Writes to different regions for the same data create conflicts
    - Conflict resolution adds complexity (last-write-wins, CRDTs, application merge)
    - Strong consistency across regions is impractical (cross-region RTT = 100ms+; Paxos would add this to every write)

  When to use:
    - Global user base where latency matters (social networks, gaming, streaming)
    - Writes can be scoped to a region (user in EU only writes to EU region)
    - Application can tolerate eventual consistency or convergence

Write localization (hybrid approach):
  In practice: user data is homed to a region (user in EU → EU is authoritative for that user)
  Active-active at the region level, but each object has a home region
  Minimizes conflicts without sacrificing global read availability

---
Data Residency

What it is:
  Legal/regulatory requirement that certain data must be stored and processed within specific geographic boundaries
  "Data about EU citizens must be stored in the EU"

GDPR implications:
  EU citizen data: must be stored in EU unless adequate data protection exists in the destination country
  Data subject rights: right to erasure (must delete from all regions where data exists)
  Data processor agreements required for cross-border transfers

How to enforce in system design:

  Geographic sharding by user home region:
    User signup → determine jurisdiction → assign to correct regional cluster
    user_id encodes region: EU users get IDs in eu-* namespace
    All writes for that user route to the EU cluster
    Read replication to other regions: only allowed for data types exempt from residency rules

  Metadata separation:
    Separate which fields are subject to residency requirements vs globally replicable
    PII (name, email, address) → stays in home region
    Anonymous analytics → can be replicated globally

  Complexity in practice:
    Multi-national users (moved from EU to US) → home region migration needed
    Backups and disaster recovery must also respect residency (backup in non-EU AZ = violation)
    Third-party services (CDN, logging, APM) must be evaluated for data they capture

---
GeoDNS

What it is:
  DNS resolution that returns different IP addresses (or CNAME records) based on the geographic location of the DNS resolver making the request

How it routes users:
  1. User's device queries their ISP's DNS resolver (or 8.8.8.8)
  2. The resolver queries your authoritative DNS (e.g. Route 53, Cloudflare)
  3. GeoDNS detects the resolver's IP → maps to a geographic region
  4. Returns the IP of the nearest load balancer in that region
  5. User connects to the regional cluster

Latency-based routing (AWS Route 53):
  Not geographic proximity but actual measured latency from AWS monitoring
  "Route this user to the region with the lowest measured RTT from their resolver"
  More accurate than pure geography (fiber routes vary; US West Coast → Asia may be faster than → East Coast)

Limitations:
  Based on resolver IP, not user IP — users on a remote VPN or corporate DNS get routed wrong
  DNS TTL: if you set TTL=60s, a region failover takes ~60s to propagate for new lookups
  Existing TCP connections are not rerouted when DNS changes

Health checks:
  Route 53 health checks can automatically remove unhealthy regional endpoints from DNS
  Combine GeoDNS + health checks for automatic failover with geographic preference

---
Cross-Region Replication

Synchronous replication:
  Write is not acknowledged until it is durably written in all target regions
  Guarantees zero data loss (RPO = 0)
  Cost: write latency = cross-region RTT (~100ms cross-continent)
  Practical limit: used only for critical financial or compliance data; unacceptable for most web apps

Asynchronous replication:
  Write is acknowledged when committed in the primary region; replication to other regions happens async
  Write latency: local only (~1ms within region)
  RPO (Recovery Point Objective): data potentially lost on primary failure = replication lag
  Typical lag: seconds to minutes depending on write volume and network conditions
  Most distributed systems use async replication for cross-region

RPO vs RTO:
  RPO (Recovery Point Objective): how much data can you afford to lose? (time window)
    RPO = 0: zero data loss → requires synchronous replication
    RPO = 5 min: acceptable to lose last 5 min of writes → async replication is fine
  RTO (Recovery Point Objective): how long can the system be down during failover?
    RTO = 0: zero downtime → requires active-active (standby ready to serve immediately)
    RTO = 30 min: manual failover process acceptable
  Lower RPO/RTO = higher cost and complexity

---
Conflict Resolution in Active-Active

Why conflicts occur:
  Two users (or the same user via two regions) write to the same key concurrently
  Both regions accept the write; replication brings both writes to each region
  Which version wins?

Last-Write-Wins (LWW):
  Mechanism: each write carries a timestamp; the write with the highest timestamp survives
  Simple and widely used (Cassandra default, S3 for versioned objects with timestamps)
  Problem: clock skew — servers in different regions can have clocks off by 10-500ms
    → a "later" write with a slightly earlier wall clock time gets incorrectly discarded
  Use when: losing concurrent updates occasionally is acceptable (social media likes, view counts)

Vector clocks / Version vectors:
  Mechanism: each value carries a version vector {regionA: 3, regionB: 2} tracking causality
  Can detect "concurrent" writes vs causally-ordered writes
  Concurrent writes → "conflict" state → must be resolved by application or user
  Used by: Amazon Dynamo (shopping cart), Riak
  Tradeoff: complexity grows with number of replicas; application must handle conflict state

CRDTs — Conflict-free Replicated Data Types:
  Data structures where any concurrent updates can be merged deterministically without conflicts
  Examples:
    G-Counter: grow-only counter — each region has its own counter; total = sum of all
    OR-Set: observed-remove set — adds and removes tracked with unique IDs
    LWW-Register: single value, last-write-wins with causal tracking
  Use when: the data type naturally supports merge (counters, sets, maps)
  Limitation: not all data models fit CRDT semantics (e.g. bank balance with overdraft prevention)

Application-level merge:
  Conflict is surfaced to the application (or user) to resolve
  Dynamo shopping cart: show both versions to the user, let them select / merge
  Google Docs: operational transforms (OT) or CRDT-based character-level merging
  Appropriate when: the application has domain knowledge that determines the correct merged state

---
Latency Numbers

Within a data center (same AZ): ~0.1-1ms
Between AZs in the same region: ~1-3ms
Cross-region within a continent (e.g. us-east-1 ↔ us-west-2): ~50-70ms
Cross-continent (e.g. US ↔ Europe): ~80-120ms RTT
Cross-continent (e.g. US ↔ Asia Pacific): ~150-200ms RTT
Geostationary satellite: ~600ms RTT (high latency, avoid for real-time)
Low-earth orbit (Starlink): ~20-40ms RTT (approaching fiber)

Implications for design:
  Synchronous cross-region write: adds 80-200ms to write latency — unacceptable for user-facing writes
  Within-region synchronous replication (3 AZs): adds ~3ms — acceptable
  Cross-region reads: fine if cached at CDN; unacceptable if uncached for every page load
  Consensus (Paxos/Raft) across 5 global regions: ~200ms per round trip — impractical for writes

---
Failover

Automatic vs manual:
  Automatic: health check detects failure → DNS update or LB update happens without human intervention
    Faster RTO (seconds to minutes); risk of "split-brain" if health check is flapping
  Manual: human reviews alerts, confirms failure, executes runbook to promote standby
    Safer for ambiguous failures; slower RTO (minutes to hours)

DNS TTL impact on failover speed:
  DNS records have a TTL; clients cache the IP for that duration
  TTL = 300s (5 min): failover takes up to 5 min for all clients to see new IP
  TTL = 60s: faster failover but more DNS query load (every 60s per client)
  TTL = 0: no caching, always fresh, but not widely supported and hammers DNS servers
  Best practice: set low TTL (60s) before a planned failover; keep normal TTL (300s) otherwise
  Note: some clients and resolvers ignore low TTLs or cache beyond stated TTL anyway

Failover anatomy (Route 53 + active-passive):
  1. Route 53 health check fails on primary region endpoint
  2. After N consecutive failures (e.g. 3 × 30s intervals = 90s), Route 53 marks it unhealthy
  3. DNS response for the domain switches to standby region IP
  4. Clients with cached DNS see old IP for up to TTL seconds
  5. Standby receives traffic; may need to catch up on replication lag (RPO window)

---
Global Load Balancing

Anycast:
  Multiple servers in different regions announce the same IP address via BGP
  Router network automatically routes client packets to the "closest" server announcing that IP
  (Closest = fewest BGP hops, not necessarily geographic distance)
  Example: Cloudflare uses Anycast for its entire global network — 1.1.1.1 is the same IP worldwide
  Benefit: automatic routing without DNS TTL delays; built into the network layer
  Limitation: no application-level routing logic; all anycast nodes must handle any request

BGP routing:
  Border Gateway Protocol — the routing protocol of the internet
  Autonomous Systems (AS) exchange routing information; each chooses best path to destination
  Anycast leverages BGP: your AS announces a prefix from multiple PoPs; internet routes to nearest PoP
  BGP convergence time: seconds to minutes — slower than DNS TTL changes for failover

Global load balancer (e.g. AWS Global Accelerator, GCP Premium Tier):
  Uses Anycast to route traffic to the nearest AWS/GCP PoP on the AWS backbone (not the public internet)
  From the PoP, traffic travels over the private backbone to the target region
  Benefit: avoids unpredictable public internet routing; consistent low latency
  Different from CloudFront: CloudFront caches content; Global Accelerator just routes to your origin

GeoDNS vs Anycast comparison:
  GeoDNS:   DNS resolution → different IP per region → works at application layer, can use health checks
  Anycast:  same IP → BGP routing → network layer, no DNS TTL delays, faster failover
  Combined: many systems use both (Anycast to a PoP, then GeoDNS/load balancer to route to correct backend)
