## Networking — Interview Questions

These are open-ended questions as they would appear in a real system design interview.
Answer each out loud or in writing before checking your notes.

---

### Load Balancing

1. What is the difference between an L4 and L7 load balancer? For each, give one use case where it is the right choice and one where it is wrong.

2. You have a fleet of application servers, but some requests take 50ms and others take 5 seconds (large file uploads). Which load balancing algorithm do you choose and why? What goes wrong if you use round-robin instead?

3. What is consistent hashing in the context of load balancing? Give a concrete scenario where it is the correct algorithm and explain what breaks if you use IP hash instead.

4. You need to do A/B testing — 10% of users should always hit the "v2" server pool, and the same user must always land in the same pool. Which load balancing algorithm enables this, and what load balancer layer (L4 or L7) is required?

5. A load balancer is sending traffic to a server that is returning 503 errors. How does health checking work, and what is the difference between a liveness check and a readiness check?

6. Your backend servers are heterogeneous — some have 32 cores and some have 8 cores. Which load balancing algorithm handles this correctly, and why does round-robin fail here?

---

### CDN & Edge

7. What types of content should you put in a CDN? What types should you never put in a CDN? Give at least three examples of each.

8. A user reports that they are seeing a stale version of your homepage 30 minutes after you deployed a change. What CDN mechanism is responsible, and what are the three strategies to solve this?

9. You deploy a critical bug fix to your JavaScript bundle. All users are seeing the broken version. You need users to get the fix immediately. Walk through each CDN invalidation strategy and explain which one works fastest and why.

10. A CDN edge node has a cache miss for an asset. Walk through the full sequence: what happens, what latency does the user experience, and what gets cached afterward?

11. What is cache versioning (fingerprinting)? Give a concrete example of how a build system implements it. Why is it considered better than explicit purges?

---

### DNS

12. Walk through the full DNS resolution process for `api.example.com` from the moment a user's browser makes the request. Name every type of server involved.

13. What is a DNS TTL and what is the trade-off between a short TTL (60 seconds) and a long TTL (3600 seconds)? When would you choose each?

14. What is GeoDNS? Give a concrete use case where it is the right tool, and explain one limitation that makes it less reliable than Anycast for latency-sensitive routing.

15. Your company runs two datacenters: US-East and EU-West. You want EU users to hit EU-West and US users to hit US-East. Walk through how GeoDNS achieves this, and what happens during a DNS failover when US-East goes down.

16. What is Anycast? How does it differ from GeoDNS, and what specific DDoS mitigation property does it provide?

---

### TCP vs UDP

17. What is the fundamental trade-off between TCP and UDP? Give a concrete use case where TCP is required and one where UDP is the right choice despite packet loss.

18. HTTP/2 runs over TCP. HTTP/3 runs over QUIC (which runs over UDP). What problem with TCP's head-of-line blocking does HTTP/3 solve, and why does this matter for web page load times?

19. A WebSocket connection is dropped during a network hiccup. What happens from the client's perspective, and what must the client implement to reconnect and resume?

20. You are designing a multiplayer game that must deliver game state updates 60 times per second. The updates are 200 bytes each. Do you use TCP or UDP, and what is the impact of that choice on how you handle packet loss?

---

### Real-world Scenarios

21. Trace a request from a user's browser to a database and back. Name every networking component it passes through, in order, and state what each component does.

22. Your application servers are in US-East. A user in Tokyo opens your app. What is the approximate RTT they experience for an uncached API call? What infrastructure change reduces this latency?

23. You are designing a global e-commerce site. During a peak traffic event, your origin servers are overwhelmed. Walk through the full set of networking layers where you can absorb or deflect traffic before it reaches the origin.

24. A CDN serves static assets and your API gateway serves dynamic requests. Both sit behind different DNS records. An engineer proposes combining them behind a single L7 load balancer that routes by URL path. What does this design look like and what are the advantages?

25. An interviewer asks: "Your service is getting 10x its normal traffic. Walk me through how you handle this from a networking perspective." What is your structured answer, covering at least three layers of the stack?
