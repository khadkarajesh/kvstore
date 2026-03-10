Real-Time Data Delivery — FAANG Interview Reference

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. When Real-Time Is Needed
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Real-time = data delivered to client within seconds of an event, without client polling.

  Use cases:
    → chat (WhatsApp, Slack) — message must appear within 1s
    → live notifications (likes, mentions) — acceptable delay ~2-5s
    → stock prices / trading — millisecond latency required
    → collaborative editing (Google Docs) — sub-100ms to feel synchronous
    → multiplayer games — sub-50ms or game feels broken
    → ride-sharing driver location — 1-2s update interval

  Key clarification question in interviews:
    "Is this server → client only, or do clients also push data to the server?"
    → server-only push → SSE is simpler
    → bidirectional → WebSockets

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
2. Short Polling
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Mechanism:
  Client sends HTTP GET every N seconds.
  Server responds immediately with current data (or empty response).

  Pros:
    → trivially simple to implement
    → stateless — every request is independent
    → works behind any proxy, firewall, CDN

  Cons:
    → most polls return no new data — wasted bandwidth and CPU
    → latency = up to N seconds (not real-time)
    → server load scales with (number of clients × poll frequency)
    → 10,000 clients polling every 5s = 2,000 req/s of mostly empty responses

  When to use:
    → acceptable latency > 30 seconds
    → very simple client (embedded device, legacy browser)
    → low number of clients

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
3. Long Polling
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Mechanism:
  Client sends HTTP request. Server holds connection open (does not respond immediately).
  When data is available → server responds. Client immediately sends next request.

  Timeline:
    Client: GET /events ────────────────────────→
    Server: (holds 28s)──────────── event → 200 response
    Client: GET /events ────────────────────────→  (immediately after response)

  Pros:
    → near-real-time delivery (latency = event latency, not poll interval)
    → works everywhere (plain HTTP, no protocol upgrade)
    → simpler than WebSockets

  Cons:
    → server holds many open connections simultaneously (thread or FD per connection)
    → not bidirectional — client cannot push data over the held connection
    → connection timeout management adds complexity
    → not efficient for very high message rates (open/close per message)

  When to use:
    → infrequent updates (notifications, job completion)
    → environments where WebSockets are blocked (corporate proxies)
    → simple server stack that cannot handle WebSocket upgrades

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
4. Server-Sent Events (SSE)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Mechanism:
  Client opens a persistent HTTP connection. Server pushes text/event-stream messages.
  Connection stays open indefinitely. One-directional: server → client only.

  HTTP response:
    Content-Type: text/event-stream
    data: {"msg": "hello"}\n\n
    data: {"msg": "world"}\n\n

  Client (browser):
    const es = new EventSource('/events');
    es.onmessage = e => console.log(e.data);

  Pros:
    → simpler than WebSockets — plain HTTP, no protocol upgrade
    → auto-reconnect built into browser EventSource API
    → works through HTTP/2 multiplexing (multiple SSE streams over one connection)
    → proxies and firewalls handle it correctly (it's just HTTP)

  Cons:
    → one-directional only — client cannot send data over the SSE connection
    → browser limit: 6 HTTP/1.1 connections per domain (mitigated by HTTP/2)
    → text-only (UTF-8) — binary data must be base64-encoded

  When to use:
    → notifications, live feeds, dashboards
    → server pushes events, client does not need to send back data
    → want simplicity over bidirectionality

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
5. WebSockets
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Mechanism:
  Persistent bidirectional TCP connection. Both sides send messages at any time.
  Established via HTTP upgrade handshake:

    Client → GET /chat HTTP/1.1
             Upgrade: websocket
             Connection: Upgrade
             Sec-WebSocket-Key: <base64>

    Server → HTTP/1.1 101 Switching Protocols
             Upgrade: websocket
             Sec-WebSocket-Accept: <derived key>

  After upgrade: raw TCP frames, no HTTP headers per message.

  Pros:
    → truly bidirectional — client and server both push freely
    → low latency (<10ms message delivery on LAN)
    → efficient — no HTTP header overhead after handshake
    → binary or text frames

  Cons:
    → stateful — server must track which socket belongs to which user
    → hard to scale horizontally (see §6)
    → some corporate firewalls and proxies block WebSocket upgrades
    → connection management overhead (heartbeats, reconnect logic)
    → load balancers need sticky sessions or L4 passthrough

  When to use:
    → chat, collaborative editing, multiplayer games, live trading
    → any scenario requiring client → server push at high frequency

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
6. Scaling WebSockets
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

The core problem:
  User A is connected to WebSocket Server 1.
  User B is connected to WebSocket Server 2.
  User A sends a message to User B.
  Server 1 has no connection to User B → cannot deliver.

Solution: pub/sub fanout layer between WebSocket servers

  Architecture:
    Server 1 receives message for user B
    → Server 1 publishes to Redis channel "user:B" (or room channel)
    → Server 2 is subscribed to "user:B"
    → Server 2 receives the published message
    → Server 2 pushes to user B's open WebSocket connection

  Implementation options:
    → Redis pub/sub: simple, in-memory, not durable (fire-and-forget)
    → Kafka: durable, ordered, replayable — use for high-throughput or audit
    → NATS: low-latency, lightweight, good for gaming/trading

---
Connection Routing

  Option A: Consistent hashing on user_id → always route same user to same server
    → predictable, avoids fan-out for direct messages
    → rebalancing on server scale-out is complex

  Option B: Any server + pub/sub fan-out (described above)
    → simpler routing, pub/sub adds ~1ms latency
    → preferred for chat and notification systems

---
Scale Numbers

  Each WebSocket server: ~10,000-100,000 concurrent connections (depends on RAM, OS limits)
  100M DAU, 10% online at peak = 10M concurrent connections → ~100-1,000 WS servers

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
7. Presence Systems (Is User Online?)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Write path:
  User connects WebSocket → SET presence:{user_id} = online EX 30  (TTL 30s in Redis)
  User sends heartbeat every 20s → SETEX presence:{user_id} 30 online  (reset TTL)
  User closes connection or app backgrounds → heartbeats stop → TTL expires → key gone

Read path:
  GET presence:{user_id} → exists? online : offline

  For bulk presence (Slack channel members):
    MGET presence:1 presence:2 presence:3 ...  (pipeline for efficiency)
    → returns list of values, null = offline

---
Edge Cases

  Mobile background → app suspended, no heartbeats → TTL expires → correctly shown offline
  Network loss without clean disconnect → no heartbeats → TTL expires → correct
  Redis failure → all users appear offline (fail-safe: default to unknown/offline)
  Clock skew between servers → irrelevant (TTL-based, not timestamp-based)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
8. Decision: Which Mechanism to Use?
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Need bidirectional (client also pushes)?     → WebSockets
  Server → client only, want simplicity?       → SSE
  Infrequent updates, WebSockets blocked?      → Long polling
  Acceptable delay > 30s, minimal complexity?  → Short polling

  Latency requirements:
    < 100ms (games, trading)  → WebSockets
    < 2s (chat, notifications) → WebSockets or SSE
    < 30s (dashboards)         → SSE or long polling
    > 30s (batch updates)      → short polling

---
Interviewer Signals

  Mentioning "short polling" without dismissing it → junior signal
  Jumping to WebSockets for every use case → missing tradeoffs
  Correct answer: name all four, give decision criteria, choose based on requirements
