## Real-Time Data Delivery — Flashcards

Format: Q → A
Cover the answer, recall it, then check.

---

**Q: What is the first clarifying question to ask when designing a real-time system?**
A: "Is this server-to-client only, or do clients also push data to the server?" Server-only push → SSE is simpler. Bidirectional → WebSockets.

---

**Q: How does short polling work?**
A: Client sends an HTTP GET every N seconds. Server responds immediately with current data or an empty response. Every request is independent and stateless.

---

**Q: What are the main drawbacks of short polling?**
A: Most polls return no new data — wasted bandwidth and CPU. Latency equals up to N seconds, so it is not truly real-time. Server load scales with (number of clients × poll frequency): 10,000 clients polling every 5s = 2,000 req/s of mostly empty responses.

---

**Q: When is short polling the right choice?**
A: When acceptable latency exceeds 30 seconds, the client is very simple (embedded device, legacy browser), or the number of clients is low.

---

**Q: How does long polling work?**
A: Client sends an HTTP request. Server holds the connection open without responding. When data is available, the server responds. The client immediately sends the next request. Latency equals the time until the event occurs, not the poll interval.

---

**Q: What are the main drawbacks of long polling?**
A: Server holds many open connections simultaneously (thread or file descriptor per connection). It is not bidirectional — the client cannot push data over the held connection. Connection timeout management adds complexity. Not efficient for high message rates because each message requires an open-and-close cycle.

---

**Q: When is long polling the right choice?**
A: Infrequent updates (notifications, job completion), environments where WebSockets are blocked by corporate proxies, or server stacks that cannot handle WebSocket upgrades.

---

**Q: How does Server-Sent Events (SSE) work?**
A: Client opens a persistent HTTP connection. Server pushes text/event-stream messages indefinitely. Communication is one-directional: server to client only. The browser's EventSource API handles auto-reconnect automatically.

---

**Q: What are the advantages of SSE over WebSockets?**
A: Simpler — plain HTTP with no protocol upgrade. Auto-reconnect is built into the EventSource API. Works through HTTP/2 multiplexing (multiple SSE streams over one TCP connection). Proxies and firewalls handle it correctly because it is just HTTP.

---

**Q: What are the limitations of SSE?**
A: One-directional only — client cannot send data over the SSE connection. Browser limit of 6 HTTP/1.1 connections per domain (mitigated by HTTP/2). Text-only (UTF-8) — binary data must be base64-encoded.

---

**Q: When is SSE the right choice?**
A: When the server pushes events and the client does not need to send back data — notifications, live feeds, dashboards. Choose SSE over WebSockets when simplicity matters more than bidirectionality.

---

**Q: How does the WebSocket handshake work?**
A: Client sends an HTTP GET with "Upgrade: websocket" and "Connection: Upgrade" headers plus a Sec-WebSocket-Key. Server responds with HTTP 101 Switching Protocols and a Sec-WebSocket-Accept header. After the upgrade, raw TCP frames are exchanged with no HTTP headers per message.

---

**Q: What are the advantages of WebSockets?**
A: Truly bidirectional — client and server both push freely at any time. Low latency (<10ms on LAN). Efficient — no HTTP header overhead after the handshake. Supports binary and text frames.

---

**Q: What are the main drawbacks of WebSockets?**
A: Stateful — server must track which socket belongs to which user, making horizontal scaling difficult. Some corporate firewalls and proxies block WebSocket upgrades. Requires sticky sessions or L4 passthrough at the load balancer. Connection management overhead (heartbeats, reconnect logic).

---

**Q: What is the core scaling problem with WebSockets and how is it solved?**
A: User A is connected to WebSocket Server 1 and User B is connected to Server 2. When A sends a message to B, Server 1 has no connection to B and cannot deliver. Solution: a pub/sub fanout layer. Server 1 publishes to a Redis channel for user B. Server 2 is subscribed to that channel, receives the message, and pushes it to B's open connection.

---

**Q: What are the options for the pub/sub fanout layer between WebSocket servers?**
A: Redis pub/sub: simple, in-memory, not durable (fire-and-forget). Kafka: durable, ordered, replayable — use for high-throughput or audit requirements. NATS: low-latency, lightweight — good for gaming and trading.

---

**Q: How do you implement a presence system (is user online)?**
A: Write path: on WebSocket connect, SET presence:{user_id} = online with TTL 30s in Redis. Client sends a heartbeat every 20s to reset the TTL. When the app closes or backgrounds, heartbeats stop and the key expires. Read path: GET presence:{user_id} — exists means online, missing means offline. For bulk presence, use MGET pipelined.

---

**Q: What happens to presence state when a mobile app is backgrounded or the network drops?**
A: Heartbeats stop. The Redis TTL expires after 30s. The user is correctly shown as offline without requiring an explicit disconnect signal. This is the key advantage of TTL-based presence over event-driven presence.

---

**Q: How many concurrent WebSocket connections can a single server handle, and what does that imply for scale?**
A: A single WebSocket server handles approximately 10,000 to 100,000 concurrent connections depending on RAM and OS limits. At 100M DAU with 10% online at peak (10M concurrent connections), you need roughly 100 to 1,000 WebSocket servers.

---

**Q: What is the decision framework for choosing between short polling, long polling, SSE, and WebSockets?**
A: Need bidirectional (client also pushes)? → WebSockets. Server-to-client only, want simplicity? → SSE. Infrequent updates, WebSockets blocked? → Long polling. Acceptable delay >30s, minimal complexity? → Short polling. By latency: <100ms (games, trading) → WebSockets; <2s (chat) → WebSockets or SSE; <30s (dashboards) → SSE or long polling; >30s → short polling.

---
