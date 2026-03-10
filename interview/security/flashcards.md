## Security — Flashcards

Format: Q → A
Cover the answer, recall it, then check.

---

**Q: What is the difference between authentication and authorization?**
A: Authentication (authn) verifies identity — who are you? Authorization (authz) verifies permission — what are you allowed to do? Authn happens first; authz is meaningless without it.

---

**Q: What are the three parts of a JWT and what does each contain?**
A: header (algorithm + token type), payload (claims: sub, exp, roles — base64-encoded, not encrypted), signature (HMAC or RSA over header.payload using a secret/private key). Payload is readable by anyone; only the signature provides integrity.

---

**Q: Why can't you revoke a JWT before it expires?**
A: JWTs are stateless — servers verify the signature without a DB lookup. There is no central record to invalidate. To revoke early you need a token blocklist (Redis set of revoked JTIs), which reintroduces statefulness.

---

**Q: What is the refresh token pattern and why does it exist?**
A: Short-lived access token (15 min) + long-lived refresh token (7-30 days). Access token used on every request — if stolen, damage window is small. Refresh token used only to get new access tokens; it can be revoked in a DB. Limits exposure without requiring frequent logins.

---

**Q: httpOnly cookie vs localStorage for storing JWTs — what is the security difference?**
A: httpOnly cookie: JS cannot read it → XSS-safe (malicious script cannot exfiltrate the token), but CSRF risk (mitigated with SameSite=Strict). localStorage: JS can read it → any XSS vulnerability can steal the token. httpOnly cookie is preferred for sensitive web apps.

---

**Q: Walk through the OAuth 2.0 authorization code flow.**
A: 1) App redirects user to auth server with client_id + scope. 2) User logs in and consents. 3) Auth server redirects back with short-lived authorization code. 4) App backend POSTs code + client_secret to token endpoint. 5) Auth server returns access_token + refresh_token. The code is browser-visible; the tokens are exchanged server-side and never touch the URL.

---

**Q: Why does OAuth use a code exchange step instead of returning the token directly in the redirect?**
A: The redirect URL is visible in browser history, server logs, and Referer headers. Sending a short-lived code instead of the actual token limits exposure. The code is then exchanged server-side using the client_secret, which never leaves the backend.

---

**Q: API key vs JWT — when do you choose each?**
A: API key: machine-to-machine, external developer integrations, long-lived service accounts, easy explicit revocation (just delete the DB record). JWT: user sessions, stateless horizontal scaling, when you need claims embedded in the token, short-lived access control.

---

**Q: How do you rate limit by API key using Redis?**
A: INCR a key scoped to the API key + time window (e.g. apikey:abc:2024-03-10:14 for the current hour). Set expiry = window duration. If count exceeds limit, reject. INCR is atomic — safe under concurrent requests without locks.

---

**Q: Where should TLS be terminated in a typical web architecture, and what is the trade-off?**
A: At the load balancer: offloads crypto from app servers, centralizes cert management, internal traffic is plaintext. At the origin: end-to-end encryption, but cert management complexity scales with fleet size and LB cannot inspect traffic. Most systems terminate at the LB and rely on VPC/private subnet for internal traffic security.

---

**Q: What is mTLS and when is it used?**
A: Mutual TLS — both client and server present and verify certificates, proving identity in both directions. Used for service-to-service communication in zero-trust internal networks. Service meshes (Istio, Linkerd) enforce mTLS transparently without application code changes.

---

**Q: What is SQL injection and what is the correct mitigation?**
A: Attacker injects SQL into user input, altering the executed query (e.g. ' OR '1'='1). Mitigation: parameterized queries / prepared statements — the query structure is fixed at compile time; user input is always treated as data, never as SQL.

---

**Q: What is XSS and what are two mitigations?**
A: Cross-Site Scripting — attacker injects malicious JS into content served to other users; it runs in the victim's browser (cookie theft, keylogging). Mitigations: 1) Output encoding — escape < > & " before rendering user content as HTML. 2) Content-Security-Policy header — restricts which scripts the browser will execute.

---

**Q: What is CSRF and how does SameSite cookie attribute prevent it?**
A: Cross-Site Request Forgery — malicious site tricks the user's browser into sending an authenticated request to your site (the browser auto-includes cookies). SameSite=Strict prevents the browser from sending the cookie on cross-origin requests entirely. SameSite=Lax allows GET but blocks POST cross-origin.

---

**Q: What is SSRF and give an example of why it is dangerous.**
A: Server-Side Request Forgery — attacker tricks the server into making HTTP requests to internal resources. Example: attacker sets a URL parameter to http://169.254.169.254/latest/meta-data/ (AWS instance metadata) and retrieves IAM credentials. Mitigation: allowlist outbound destinations, block private IP ranges.

---

**Q: Path traversal — what is it and how do you prevent it?**
A: Attacker uses ../ sequences in a filename parameter to escape the intended directory and read arbitrary files (e.g. ../../etc/passwd). Prevention: canonicalize the resolved path and verify it starts with the expected base directory before any file operation.

---

**Q: Environment variables vs a secret manager — what does the secret manager add?**
A: Secret managers (Vault, AWS Secrets Manager) add: encryption at rest, fine-grained access control with audit logs (who read what and when), automatic rotation, and dynamic short-lived credentials. Env vars have none of these — they are visible to all process code and have no rotation or audit trail.

---

**Q: What is defense in depth?**
A: A layered security approach where multiple independent controls protect a system. No single control is the last line of defense. If a WAF is bypassed, parameterized queries still prevent SQLi. If a token is stolen, short TTL limits the damage window. Layers: network (firewall), edge (WAF), transport (HTTPS/mTLS), application (input validation, authz), data (encryption at rest).

---

**Q: A service issues JWTs signed with HS256. An attacker changes alg to "none" in the header. What happens and how do you prevent it?**
A: Some naive JWT libraries accept "none" as a valid algorithm and skip signature verification — the attacker can forge any token. Prevention: always explicitly specify and enforce the expected algorithm server-side; never accept whatever algorithm the token declares.

---

**Q: Your microservices use API keys for internal service-to-service auth. What is the upgrade and why?**
A: Move to mTLS. API keys are long-lived static secrets — if leaked, any service can impersonate another. mTLS uses short-lived certificates where the private key never leaves the service. A service mesh (Istio/Linkerd) can enforce this transparently, with automatic cert rotation.

---
