## Rate Limiting — Flashcards

Format: Q → A
Cover the answer, recall it, then check.

---

**Q: What is token bucket and what are its two key parameters?**
A: A bucket holding up to N tokens, refilled at rate R/sec. Each request consumes 1 token. If empty, reject. N = burst capacity, R = sustained throughput.

---

**Q: Does token bucket allow bursts? Why?**
A: Yes. Tokens accumulate when traffic is low. A burst can consume all N tokens instantly before the bucket empties and throttling begins.

---

**Q: What do you store per user in Redis to implement token bucket?**
A: (token_count, last_refill_time). On each request, compute elapsed time → add new tokens → cap at N → allow or reject.

---

**Q: What is leaky bucket?**
A: Requests enter a fixed-size queue. A worker drains the queue at a fixed rate R. If queue is full, incoming requests are rejected.

---

**Q: What is the key difference between token bucket and leaky bucket?**
A: Token bucket allows bursts up to N tokens. Leaky bucket never allows bursts — output is always at constant rate R regardless of input.

---

**Q: What is fixed window counter?**
A: Divide time into fixed windows. Count requests per window using a Redis key with TTL. If count exceeds limit, reject. Simple but has boundary burst problem.

---

**Q: What is the boundary burst problem in fixed window counter?**
A: 100 req/min limit. 100 requests at 0:59 (window 1) + 100 requests at 1:01 (window 2) = 200 requests in 2 seconds, both windows allowed it.

---

**Q: What is sliding window log and why is it memory-expensive?**
A: Store timestamp of every request in a sorted set. Remove old timestamps, count remaining, reject if over limit. Accurate but stores every request timestamp — scales with traffic volume.

---

**Q: What is the sliding window counter formula?**
A: estimated_count = current_window_count + (previous_window_count × overlap_fraction), where overlap_fraction = portion of previous window within the rolling window.

---

**Q: Why is sliding window counter preferred over sliding window log in production?**
A: Only stores 2 counters per user (not every timestamp). Accurate enough — Cloudflare measured <0.003% error. Best balance of memory, accuracy, and simplicity.

---

**Q: What are the three approaches to distributed rate limiting and their trade-offs?**
A: 1) Centralized Redis — accurate, adds ~500µs latency, Redis is single point of failure. 2) Local counters + periodic sync — fast, allows brief overflow during sync interval. 3) Sticky sessions — simple but uneven load, state lost on server failure.

---

**Q: Which distributed rate limiting strategy is best for most production APIs?**
A: Local counters + periodic sync. Brief overflow of a few percent is acceptable for most APIs, and it avoids the Redis latency on every request.

---

**Q: What HTTP status code does rate limiting return?**
A: 429 Too Many Requests

---

**Q: What are the four standard rate limit response headers?**
A: X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset, Retry-After

---

**Q: Where is the most common place to enforce rate limiting in a system architecture?**
A: API Gateway — applied before requests reach any backend service, so limits are enforced in one place across all services.

---

**Q: Is client-side rate limiting a security mechanism? Why or why not?**
A: No. Untrusted clients can ignore it. It is a courtesy mechanism to reduce unnecessary requests, not a protection against malicious actors.

---

**Q: You need strict per-user rate limiting with no allowed overflow. Which algorithm do you choose?**
A: Sliding window log (exact) or centralized Redis token bucket. Both are accurate but expensive — justify with the strictness requirement.

---

**Q: You need rate limiting that smooths traffic to protect a downstream payment service. Which algorithm?**
A: Leaky bucket — constant output rate prevents any burst from reaching the downstream service.

---

**Q: Token bucket with N=100, R=10/sec. User sends 100 requests instantly, then 5 requests 3 seconds later. How many are allowed?**
A: All 100 instant requests allowed (bucket drains to 0). After 3 seconds: 30 new tokens added → all 5 allowed. 25 tokens remain.

---

**Q: What should a well-behaved client do when it receives a 429?**
A: Read the Retry-After header and wait that duration. Use exponential backoff with jitter on retries. Never immediately retry in a loop.
