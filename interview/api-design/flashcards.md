# API Design — Flashcards

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

**Q: What are the core principles of REST?**
A: Stateless (each request self-contained), resource-based URLs (nouns not verbs), HTTP verbs carry semantics (GET=read, POST=create, PUT=replace, DELETE=remove), uniform interface.

---

**Q: What are REST's two biggest weaknesses at scale?**
A: Over-fetching (endpoint returns more data than client needs) and under-fetching (client needs multiple endpoints to assemble one view → multiple round trips).

---

**Q: What are gRPC's advantages over REST?**
A: Binary Protocol Buffers (smaller payload than JSON), HTTP/2 (multiplexing, header compression), strongly typed schema, generated clients for all languages, native streaming support (server, client, bidirectional).

---

**Q: What are gRPC's weaknesses?**
A: Not human readable (binary, hard to debug without tooling), limited browser support (gRPC-Web needs a proxy), requires .proto schema to be shared and versioned.

---

**Q: What are the four gRPC communication patterns?**
A: Unary (1 request → 1 response), server streaming (1 request → many responses), client streaming (many requests → 1 response), bidirectional streaming (both sides stream simultaneously).

---

**Q: When would you choose GraphQL over REST?**
A: Complex data graphs with many relationships (social network), mobile clients constrained by bandwidth, rapid frontend iteration where frontend changes shape of data without backend changes.

---

**Q: What is the N+1 query problem in GraphQL?**
A: Fetching a list of N users, then for each user fetching their posts → N+1 database queries. Fix: DataLoader, which batches all user-ID lookups into one query per resolver level.

---

**Q: Why is caching harder with GraphQL than REST?**
A: All GraphQL operations typically use POST requests, not GET, so HTTP-level caching (CDN, browser cache) does not apply automatically. REST GET requests are cacheable by URL.

---

**Q: REST vs gRPC vs GraphQL — the decision rule.**
A: External public API → REST. Internal microservices → gRPC. Complex relational data or mobile bandwidth-constrained clients → GraphQL. Real-time bidirectional → WebSockets.

---

**Q: What is the key weakness of offset-based pagination?**
A: Two problems: (1) unstable — a new insert shifts all subsequent pages, causing duplicates or skips; (2) slow at scale — database must scan and discard all rows before the offset, O(offset) cost.

---

**Q: Why is cursor-based pagination better at scale?**
A: Stable: cursor points to a specific row, not a count — concurrent inserts don't shift pages. Fast: database uses an index seek directly to the cursor position, O(log N) regardless of offset size.

---

**Q: What does a cursor in cursor-based pagination encode?**
A: A position in the result set — typically the ID or timestamp of the last item returned, often base64-encoded and opaque to the client. Server uses it in a WHERE clause: WHERE id > :cursor.

---

**Q: Should you use URL versioning or header versioning for a public API?**
A: URL versioning (/api/v1/, /api/v2/). It is visible, discoverable, easy to test in browser, and CDN-cacheable. Header versioning is cleaner but harder for external consumers to use.

---

**Q: What problem does an idempotency key solve?**
A: Client sends a mutation (e.g. payment), network times out, client cannot know if it succeeded. Without idempotency key, retry causes double execution (double charge). With the key, server returns cached result on retry.

---

**Q: How do you implement idempotency keys?**
A: Client generates UUID, sends with request. Server stores (key → response) in Redis with TTL. On first request: execute and store. On duplicate key: return stored response without re-executing.

---

**Q: Which mutation types require idempotency keys?**
A: Any mutation with side effects that must not be duplicated: payments, order placement, email sends, notification dispatches, inventory decrements.

---

**Q: What is the senior interview script for API design?**
A: "For the client-facing API I would use REST with cursor-based pagination. For internal services I would use gRPC. Every mutation endpoint requires an idempotency key stored in Redis to handle retries safely."

---

**Q: What HTTP verb is idempotent but NOT safe?**
A: DELETE and PUT. Safe means no side effects (GET). Idempotent means calling it N times has the same effect as calling it once. DELETE is idempotent (deleting already-deleted resource → still gone) but not safe.

---

**Q: What is WebSockets used for and when does REST/gRPC not suffice?**
A: WebSockets provide persistent, full-duplex, bidirectional TCP connections. Use when server must push data to client without client polling: live chat, real-time dashboards, multiplayer games, live price feeds.

---

**Q: What is over-fetching and why does it matter on mobile?**
A: Over-fetching: API returns more fields than the client needs. On mobile, every extra byte costs battery (radio transmission) and user data. GraphQL solves this by letting the client specify exactly which fields to return.

---
