# Networking — Flashcards

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

**Q: What is the key difference between L4 and L7 load balancers?**
A: L4 routes by IP + port only (no content inspection, fast, TCP-level). L7 routes by HTTP headers, URL path, cookies (content-aware, slower, supports sticky sessions and SSL termination).

---

**Q: When would you choose L4 over L7?**
A: Raw TCP traffic (databases, game servers), maximum throughput with minimum CPU overhead, non-HTTP protocols.

---

**Q: When would you choose L7 over L4?**
A: HTTP services needing path-based routing, SSL/TLS termination, sticky sessions, A/B routing, or WebSocket upgrade handling.

---

**Q: What is the weakness of round-robin load balancing?**
A: It distributes requests by count, not by server load. If some requests take much longer than others, slow servers accumulate connections while others are idle. Use least-connections for variable request durations.

---

**Q: Why is consistent hashing better than modulo hashing for load balancing a cache cluster?**
A: Modulo hashing: adding or removing 1 server remaps nearly all keys → massive cache invalidation. Consistent hashing: adding/removing 1 server remaps only ~1/N of keys, preserving cache affinity for the rest.

---

**Q: What is the use case for consistent hashing in load balancing?**
A: Any stateful service requiring affinity: distributed caches (same key always hits same node), session stores, sharded databases. Ensures same user/key always routes to the same backend.

---

**Q: What should you cache in a CDN?**
A: Static assets (images, CSS, JS bundles, videos, fonts), infrequently-changing public API responses, large file downloads.

---

**Q: What should you NOT cache in a CDN?**
A: User-specific data (cart, profile), real-time data (live prices, scores), mutations (POST/PUT/DELETE), responses requiring per-request auth checks.

---

**Q: What are the three CDN invalidation strategies?**
A: (1) TTL expiry — wait for TTL, simple but slow. (2) Explicit purge — call CDN API, fast but operationally heavy. (3) Cache versioning — change URL when content changes (main.v2.js), no invalidation needed, best practice.

---

**Q: Why is cache versioning (URL fingerprinting) the best CDN invalidation strategy?**
A: The URL changes when content changes (e.g. build hash in filename). Old URL is still cached harmlessly. New URL is fetched fresh automatically. Zero operational overhead, zero risk of stale content.

---

**Q: What is GeoDNS and what problem does it solve?**
A: DNS returns different IP addresses based on the client's geographic location. Solves: reducing latency by routing users to the nearest datacenter, and data residency compliance (e.g. EU users → EU servers).

---

**Q: What is Anycast and how does it differ from GeoDNS?**
A: Anycast: one IP advertised from multiple datacenters; BGP routes each packet to the topologically nearest one. GeoDNS: different IPs per region, routing at DNS level. Anycast is faster to failover (no DNS TTL), operates at network layer.

---

**Q: Name three use cases for Anycast.**
A: CDNs (Cloudflare, Fastly), DNS providers (1.1.1.1, 8.8.8.8 are Anycast IPs), DDoS mitigation (attack traffic spread across all PoPs).

---

**Q: What is a health check in the context of load balancing?**
A: A periodic probe the load balancer sends to each backend to verify it is alive and ready. Failing backends are removed from rotation. Two types: liveness (is it alive?) and readiness (is it ready for traffic?).

---

**Q: Describe the typical request flow from user to database in a standard architecture.**
A: User → DNS → CDN edge (static asset cache) → L7 Load Balancer (HTTP routing, SSL termination) → Application Services → [L4 Load Balancer if needed] → Databases/Caches.

---
