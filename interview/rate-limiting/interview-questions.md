## Rate Limiting — Interview Questions

These are open-ended questions as they would appear in a real system design interview.
Answer each out loud or in writing before checking your notes.

---

### Algorithms

1. Walk through how the Token Bucket algorithm works. What does the bucket hold, what does "replenishment" mean, and what happens when the bucket is empty?

2. What is the key difference between a Token Bucket and a Leaky Bucket? In which scenario is each the correct choice?

3. A payment processor cannot handle burst traffic — it requires a strictly smooth input rate. Which rate limiting algorithm do you choose and why? What happens if you used Token Bucket instead?

4. What is the boundary burst problem with Fixed Window Counter? Give a specific numerical example showing how a user can send 2× the limit in 2 seconds without being rejected.

5. Compare Sliding Window Log and Sliding Window Counter (hybrid). What does each trade off, and which do you recommend for a production system serving 50M users? Why?

6. Cloudflare uses the Sliding Window Counter. What approximation does it make, and how accurate is it in practice?

---

### Distributed Rate Limiting

7. You have 10 application servers, each with a local rate limiter set to 100 req/min per user. What is the actual effective limit a user can reach across the fleet?

8. Walk through the centralized Redis counter approach to distributed rate limiting. What is the latency cost, and what happens if Redis goes down?

9. Walk through the local counter with periodic sync approach. What is the trade-off, and under what conditions is "brief overflow" acceptable?

10. An engineer proposes sticky sessions: always route the same user to the same server, so local counters are accurate. What are two failure modes of this approach?

11. You are designing a rate limiter for a globally distributed API (users in US, EU, Asia). Describe the trade-off between a global centralized counter, per-region counters, and local per-server counters. Which do you recommend?

---

### Implementation Trade-offs

12. Where does rate limiting live in the architecture — API gateway, service layer, or client side? Give a reason to use each location, and explain why client-side rate limiting is not a security control.

13. A rate limiter uses Redis with `INCR` and a TTL. Write the pseudocode for a Fixed Window Counter check. What is the race condition and how do you eliminate it?

14. Your API returns HTTP 429. What four headers do you include in the response, and why does each matter for a well-behaved client?

15. You need to implement per-user, per-IP, and global rate limits on the same endpoint. How do you structure the Redis keys, and in what order do you check them?

16. An engineer argues that rate limiting should be implemented in each microservice rather than at the API gateway. What are two advantages and two disadvantages of this approach?

---

### Failure Modes

17. Your Redis rate limiter cluster goes down. What should your rate limiting layer do — fail open (allow all traffic) or fail closed (reject all traffic)? What are the consequences of each choice?

18. A legitimate user has a background job that fires 500 requests in 1 second at startup, then goes quiet. They are hitting your 100 req/min limit and getting 429s. How do you handle this without changing the limit for everyone?

19. A single client is sending 10,000 requests per second to your API, bypassing the 1,000 req/min limit by using 20 different IP addresses. What rate limiting dimension catches this, and how?

20. A retry storm begins: clients receive 429 responses, immediately retry, generating more 429s, which cause more retries. How do you break this cycle from both the server side and the client side?

---

### Design Problems

21. Design the rate limiting system for a public REST API like Twitter's. What dimensions do you rate limit on, which algorithm do you use, where does the limiter live, and how do you handle the distributed nature of the fleet?

22. You are designing rate limiting for a payment API that must allow bursts for legitimate retry behavior but protect against runaway clients. Walk through your algorithm choice, configuration, and what headers you return on 429.

23. A new feature launches and generates 50× the expected traffic from a single enterprise customer. Walk through how your rate limiting system detects this, what response the customer receives, and how you work with them to raise their limit safely.

24. Design a rate limiter that applies different limits to different customer tiers: free (100 req/min), pro (10,000 req/min), enterprise (unlimited with fair-use monitoring). How do you structure this in Redis and at the API gateway?

25. An interviewer says "add rate limiting to this design." Give the complete answer: algorithm, storage, location in architecture, failure mode handling, and what you return to the client when the limit is exceeded.
