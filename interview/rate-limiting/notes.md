## Rate Limiting

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

### What Rate Limiting Is and Why It Exists

  Goal: protect services from abuse, DoS, cost overruns, and unfair resource consumption.

  Types of limits:
  - User-level    → per authenticated user ID
  - IP-level      → per client IP (catches unauthenticated abuse)
  - API key-level → per developer key (tiered pricing / quotas)
  - Global        → protect an entire endpoint regardless of who calls it

  Without rate limiting:
  - One bad actor (or buggy client) can saturate your service
  - Downstream databases get overwhelmed during traffic spikes
  - Cloud costs become unbounded

---

### Token Bucket

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Mechanism:
  - Bucket holds max N tokens
  - Tokens added at rate R per second (replenishment)
  - Each request consumes 1 token
  - If bucket empty → reject request (HTTP 429)

  Key property: allows bursts up to bucket size N, then throttles at rate R.

  Implementation (per user in Redis):
    store: (token_count, last_refill_time)
    on request:
      elapsed = now - last_refill_time
      new_tokens = elapsed × R
      token_count = min(token_count + new_tokens, N)
      if token_count >= 1 → allow, decrement token_count
      else → reject

  Used by: AWS API Gateway, Stripe, most major API platforms.

  When to use: APIs where occasional bursts are acceptable.

---

### Leaky Bucket

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Mechanism:
  - Requests enter a queue (the "bucket")
  - Queue is drained at a fixed rate R (a worker processes one request per 1/R seconds)
  - If queue is full → reject incoming request

  Key property: output rate is strictly constant — no bursts leak through.

  Difference from token bucket:
  - Token bucket → allows bursts up to N, then limits
  - Leaky bucket → never allows bursts; output is always smooth

  When to use: protecting a downstream service that cannot handle spikes
  (e.g., a payment processor, a rate-sensitive third-party API).

---

### Fixed Window Counter

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Mechanism:
  - Divide time into fixed windows (e.g. each 1-minute window has its own counter)
  - Each request increments the counter for the current window
  - If counter > limit → reject

  Implementation (Redis):
    key = "user:{id}:{window_start}"
    INCR key → if result > limit, reject
    SET key TTL = window_size (so old keys auto-expire)

  Problem — boundary burst:
    Limit = 100 req/min
    100 requests arrive at 0:59 (last second of window 1) → allowed
    100 requests arrive at 1:01 (first second of window 2) → allowed
    → 200 requests in 2 seconds despite 100/min limit

  When to use: simple use cases where boundary burst is acceptable.
  Not appropriate for strict enforcement.

---

### Sliding Window Log

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Mechanism:
  - Store timestamp of every request in a sorted set per user
  - On each request:
      remove all timestamps older than (now - window_size)
      count remaining timestamps
      if count >= limit → reject
      else → add current timestamp, allow

  Key property: perfectly accurate, no boundary burst problem.

  Cost: memory scales with request volume — stores every request timestamp.
  1M users × 100 requests in-window = 100M timestamps stored.

  When to use: low-volume, high-precision scenarios where accuracy > memory cost.

---

### Sliding Window Counter (Hybrid)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Mechanism:
  - Track counts for: current window + previous window
  - Approximate current rate:
      estimated_count = current_window_count + (previous_window_count × overlap_fraction)
      overlap_fraction = fraction of the previous window that falls within the rolling window

  Example:
    Window = 1 min. Current time = 0:45 of the current minute.
    overlap_fraction = (60 - 45) / 60 = 0.25
    estimated_count = current_count + previous_count × 0.25

  Key properties:
  - Approximation only — slightly over or under counts at window boundaries
  - In practice: accurate enough for production, Cloudflare measured <0.003% error
  - Memory: only 2 counters per user per window — extremely cheap

  Used by: Cloudflare, most production-grade rate limiters.
  When to use: default choice. Best balance of accuracy, memory, and simplicity.

---

### Distributed Rate Limiting

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Problem: rate limiter on server A has no knowledge of requests hitting server B.
  Without coordination → each server allows up to the limit → total traffic = N × limit.

  Solution 1 — Centralized Redis counter:
    All servers read/write a single Redis counter per user.
    → Accurate: every request hits the same counter
    → Cost: one Redis round-trip per request (~500 µs same-datacenter), Redis becomes hot path
    → Single point of failure if Redis goes down (mitigate with Redis Cluster)

  Solution 2 — Local counters + periodic sync:
    Each server tracks locally. Every few seconds, sync deltas to Redis.
    → Fast: no Redis hop on the critical path
    → Cost: brief overflow possible during sync interval
    → Acceptable for non-critical limits (burst allowance covers it)

  Solution 3 — Sticky sessions:
    Route same user to same server always (consistent hashing on user ID).
    → Simple: local counter is accurate for that user
    → Cost: uneven load distribution, server failure loses all state for routed users

  Interview move: state the trade-off explicitly.
    "Centralized is accurate but adds latency. Local sync is fast but allows brief overflow.
    For most APIs, brief overflow of a few percent is fine — use local + periodic sync."

---

### Where Rate Limiting Lives

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  API Gateway (most common):
  - Rate limiting applied before requests reach any service
  - One place to configure and enforce limits across all downstream services
  - Examples: Kong, AWS API Gateway, Nginx, Envoy

  Service layer:
  - Each service enforces its own limits
  - Useful for per-service quotas (e.g., search service has stricter limits than profile service)
  - More complex operationally

  Client-side:
  - Clients implement courtesy backoff / retry-after logic
  - Not a security mechanism — untrusted clients can ignore it
  - Reduces unnecessary requests to your servers

---

### Rate Limit Response

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  HTTP status: 429 Too Many Requests

  Standard headers to return:
    X-RateLimit-Limit      → the limit (e.g. 100)
    X-RateLimit-Remaining  → remaining requests in current window
    X-RateLimit-Reset      → Unix timestamp when the window resets
    Retry-After            → seconds until client may retry

  Good client behavior: exponential backoff + jitter on 429.
  Bad client behavior: immediate retry loop → makes the problem worse.
