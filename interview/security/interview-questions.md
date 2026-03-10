## Security — Interview Questions

These are open-ended questions as they would appear in a real system design interview.
Answer each out loud or in writing before checking your notes.

---

### Authentication & Tokens

1. A user logs into your web app and you issue them a JWT. Ten minutes later, the user is banned by an admin. The user still has a valid JWT with 50 minutes remaining. How do you handle this? What are the tradeoffs of each approach?

2. Your team debates: "Should we use JWTs or server-side sessions?" The app has 10 million active users and runs on 50 horizontally scaled app servers. Walk through the tradeoffs and give a recommendation.

3. A JWT payload contains {"role": "user"}. The client modifies it to {"role": "admin"} and replays it. What happens, and what prevents this attack?

4. You are designing a mobile app that authenticates users. The app stores the refresh token on the device. What are the risks and how do you mitigate them?

5. Your company uses OAuth 2.0 to let third-party apps access user data. A user revokes access for a third-party app. The third-party still has a valid access token that expires in 1 hour. How do you handle immediate revocation?

---

### Authorization & API Security

6. You are building a multi-tenant SaaS application. User A and User B are both authenticated. What prevents User A from reading User B's data, even if they guess the correct resource ID?

7. You have a public REST API. You issue API keys to each developer. A key gets leaked on GitHub. Walk through your incident response process and how you design to minimize future impact.

8. What is the difference between RBAC and ABAC? Give a concrete example where ABAC is necessary because RBAC is insufficient.

9. A user with the "viewer" role makes a POST request to a write endpoint. Your authz check is only in the frontend — the backend has no check. What is the vulnerability and what is the fix?

10. You are designing a rate limiter for your API. Some endpoints are more expensive than others (e.g. /search costs 10x more than /status). How do you implement cost-based rate limiting?

---

### Transport Security

11. A user visits http://example.com (no S) from a coffee shop. Describe the attack surface. What HTTP header prevents this in the future after the first HTTPS visit?

12. Your load balancer terminates TLS. Traffic between the load balancer and your app servers is plaintext over a private subnet. Is this acceptable? Under what conditions would you require end-to-end TLS?

13. You are designing a payment processing service that receives card data. PCI-DSS requires encryption in transit. Walk through your TLS configuration choices (versions, cipher suites, certificate management).

14. A junior engineer proposes using self-signed certificates for internal service-to-service communication. What are the risks, and what is a better alternative?

15. Your service makes an outbound HTTP request to a URL provided by a user. Describe the SSRF attack vector and walk through your defense strategy.

---

### Design Problems

16. Design an API gateway that handles authentication and authorization for 50 downstream microservices. The gateway must support: JWT-based user auth, API key auth for machine clients, rate limiting per client, and request logging. What are the key components and failure modes?

17. How would you secure service-to-service communication in a Kubernetes cluster running 20 microservices? Walk through your approach from network policies to identity and certificate management.

18. You are designing a secret rotation system for a production database password used by 30 app instances. The rotation must happen with zero downtime. Walk through the process.

19. Design a "Login with Google" feature for your web app. Cover the full OAuth 2.0 flow, what you store server-side, how you handle new vs returning users, and what happens if the user revokes Google access.

20. You are the security reviewer for a new feature: users can upload a profile image by providing a URL. The server fetches the image from that URL. Identify every security vulnerability in this design and propose mitigations.
