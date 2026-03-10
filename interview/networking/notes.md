# Networking — Senior System Design Interview Notes

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
## 1. LOAD BALANCERS — L4 vs L7
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

L4 — TRANSPORT LAYER
  → routes by IP address + port only
  → no content inspection — does not read HTTP headers or body
  → fast: low CPU overhead, high throughput
  → maintains TCP connection state
  → when to use: raw TCP (databases, game servers, anything non-HTTP), maximum throughput

L7 — APPLICATION LAYER
  → routes by HTTP headers, URL path, cookies, hostname
  → can do: sticky sessions, A/B routing, path-based routing, SSL termination
  → higher CPU overhead than L4 (must parse HTTP)
  → when to use: HTTP services, need routing logic, SSL termination, WebSocket upgrade

DECISION TABLE
  Raw TCP / database proxying → L4
  HTTP routing by path or header → L7
  SSL/TLS termination → L7
  Maximum throughput, minimal overhead → L4

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
## 2. LOAD BALANCING ALGORITHMS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

ROUND ROBIN
  → requests distributed sequentially: server 1, 2, 3, 1, 2, 3...
  → simple, even distribution by count
  → weakness: bad for long-running requests — a slow server accumulates connections
  → use when: request durations are uniform and short

LEAST CONNECTIONS
  → routes to server with fewest active connections at the moment
  → better for variable request duration
  → use when: some requests take much longer than others

CONSISTENT HASHING
  → hash(key or user_id) → always maps to the same server
  → adding/removing servers only remaps a fraction of keys (unlike modulo hashing)
  → use when: stateful services, caching (cache affinity — same user always hits same cache node)
  → critical for: distributed caches, session stores, sharded databases

IP HASH
  → hash of client IP → same backend server every time
  → simple sticky sessions
  → weakness: uneven distribution if many clients share one IP (NAT); breaks if server removed

WEIGHTED ROUND ROBIN / WEIGHTED LEAST CONNECTIONS
  → assign weights based on server capacity
  → powerful server gets more requests proportionally

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
## 3. CDN (CONTENT DELIVERY NETWORK)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

WHAT IT IS
  → globally distributed network of edge servers that cache content near users
  → user request → nearest edge PoP → cache hit → response in <10ms
  → cache miss → edge fetches from origin (100-500ms) → caches locally

WHAT TO CACHE
  → static assets: images, CSS, JS bundles, videos, fonts
  → API responses that change infrequently (product catalog, public config)
  → large file downloads

WHAT NOT TO CACHE
  → user-specific data (shopping cart, profile)
  → real-time data (stock prices, live scores)
  → POST/PUT/DELETE requests (mutations)
  → data requiring auth checks on every request

CDN INVALIDATION — THREE STRATEGIES

  TTL EXPIRY
    → every cached object has TTL (e.g. 1 hour)
    → stale content served until TTL expires
    → simple, no operational overhead
    → weakness: slow to propagate changes

  EXPLICIT PURGE
    → call CDN API to invalidate specific URL immediately
    → fast: change visible in seconds
    → weakness: operational overhead, easy to forget

  CACHE VERSIONING (best practice)
    → change the URL when content changes: main.js → main.v2.js
    → old URL still cached (harmless), new URL fetched fresh automatically
    → no invalidation needed, no stale content possible
    → use for: JS bundles, CSS files with build hashes

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
## 4. DNS AND GEODNS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

DNS (DOMAIN NAME SYSTEM)
  → translates human-readable domain → IP address
  → hierarchical: root → TLD (.com) → authoritative nameserver for domain
  → TTL: how long resolver caches the record before re-querying

GEODNS
  → authoritative nameserver returns different IP based on client's geographic location
  → user in US → US datacenter IP
  → user in Asia → Asia datacenter IP
  → use cases:
      → reduce latency: route to nearest datacenter
      → data residency: EU users must hit EU servers (GDPR compliance)
      → disaster recovery: failover to different region

LIMITATION OF GEODNS
  → based on resolver location, not always user location (corporate proxies, VPNs)
  → DNS TTL means failover is not instant (TTL must expire)
  → not a replacement for Anycast for latency-sensitive traffic

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
## 5. ANYCAST
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

WHAT IT IS
  → the same IP address is advertised from multiple datacenters simultaneously
  → BGP routing protocol directs each user's packet to the topologically nearest datacenter
  → from the user's perspective: one IP, one destination — they don't know routing happened

HOW IT DIFFERS FROM GEODNS
  → GeoDNS: different IPs per region, routing at DNS level, DNS TTL lag
  → Anycast: one IP, routing at network layer (BGP), near-instant failover

USE CASES
  → CDNs (Cloudflare, Fastly use Anycast heavily)
  → DNS providers (1.1.1.1, 8.8.8.8 are Anycast IPs)
  → DDoS mitigation: attack traffic distributed across all PoPs, no single point absorbs all traffic

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
## 6. TYPICAL REQUEST FLOW THROUGH ARCHITECTURE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

User
  → DNS lookup → GeoDNS or Anycast IP
  → CDN edge (static assets served here if cache hit)
  → L7 Load Balancer (HTTP routing, SSL termination)
  → Application Services
  → L4 Load Balancer (if needed for internal TCP to databases)
  → Databases / Caches

FULL FLOW SUMMARY
  User → CDN (static) → L7 LB → Services → [L4 LB] → Databases

HEALTH CHECKS AT EACH TIER
  → CDN health checks origin servers
  → L7 LB health checks application pods (liveness + readiness)
  → Application health checks database connections
