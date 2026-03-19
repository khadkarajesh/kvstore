# WhatsApp Messaging System

**Scenario:** Design a messaging system supporting 1-on-1 and group chats.

Requirements:
- Send and receive messages in real time
- Messages must never be lost even if recipient is offline
- Support group chats up to 1,000 members
- Show message delivery status: sent → delivered → read
- Scale: 2 billion users, 100 billion messages per day

---

## High-Level Architecture

```
Client (sender)
      ↓  WebSocket
WebSocket Server
      ↓
Message Service
      ↓
Kafka  (durability layer)
      ↓
Kafka Consumer
      ↓
Redis pub/sub  (delivery layer)
      ↓
WebSocket Server (holding recipient's connection)
      ↓
Client (recipient)

Offline path:
  Message Service → DB (store first) → Push Notification (APNs/FCM)
  → device wakes → app reconnects WebSocket → fetches from DB
```

---

## Component Breakdown

### 1. Delivery Mechanism — WebSocket (not SSE)

WhatsApp requires **bidirectional** communication:
- Server → client: incoming messages, delivery receipts
- Client → server: sending messages, read receipts, typing indicators

SSE is unidirectional (server → client only) — wrong choice here.
WebSocket is the correct choice for any chat system.

---

### 2. Message Flow — Step by Step

```
Step 1: Sender types message, hits send
Step 2: Message sent over sender's WebSocket connection to WS Server
Step 3: WS Server forwards to Message Service
Step 4: Message Service writes to DB first (persisted before delivery)
Step 5: Message Service publishes to Kafka
Step 6: Kafka Consumer reads, publishes to Redis pub/sub channel
Step 7: Redis delivers to WS Server holding recipient's connection
Step 8: WS Server writes to recipient's open WebSocket
Step 9: Recipient receives message → WS Server sends "delivered" receipt back
Step 10: Sender receives "delivered" status update
```

Critical rule — DB write happens at Step 4, BEFORE delivery attempt.
If recipient is offline and delivery fails, message is already safe in DB.
Never store to DB only on successful delivery — that loses offline messages.

---

### 3. Two Layers — Kafka + Redis

Same principle as live sports score system. They are complementary, not alternatives.

```
Kafka  (durability layer):
  - Message is never lost
  - Ordered per partition (chat ordering guaranteed)
  - Replayable if consumer crashes
  - Source of truth

Redis pub/sub  (delivery layer):
  - Sub-millisecond broadcast to WebSocket servers
  - Fire-and-forget — no persistence
  - One PUBLISH → correct WS server receives and delivers
```

Flow:
  Message Service → Kafka → Kafka Consumer → Redis PUBLISH → WS server → client

The Kafka Consumer is the bridge between the two layers.
Kafka does not talk to Redis directly.

---

### 4. Offline Message Handling

When recipient is offline, two things must happen:

  Step A — Message already in DB (from Step 4 above). Safe.

  Step B — Push notification via APNs (iOS) or FCM (Android):
    Sends a silent push to wake the device.
    App receives push → reconnects WebSocket → fetches missed messages
    from DB using last known sequence number.

  Without push notifications, the offline user's phone never knows
  a message arrived. The queue is never pulled. App stays silent.

```
Offline flow:
  Message arrives
      ↓
  Store to DB (always)
      ↓
  Redis fan-out attempt → no WS server holds this user's connection
      ↓
  Push Notification Service → APNs / FCM → device wakes up
      ↓
  App reconnects WebSocket
      ↓
  Client sends last sequence number: "I have up to seq=47"
      ↓
  Server fetches seq > 47 from DB → sends to client
      ↓
  Client marks messages as delivered
```

---

### 5. Message Database — Cassandra

Cassandra is the standard choice for chat message storage.

Schema:
  Partition key:   chat_id  (direct: sorted user IDs, group: group_id)
  Clustering key:  message_id (time-ordered, e.g. UUID v7 or Snowflake ID)
  Columns:         sender_id, content, status, created_at, media_url

Why Cassandra:
  - High write throughput (append-only, LSM tree)
  - Partition by chat_id keeps all messages for a conversation together
  - Clustering by message_id gives cheap range queries (fetch last N messages)
  - Scales horizontally across nodes

Query patterns:
  Fetch recent messages:  SELECT WHERE chat_id = ? ORDER BY message_id DESC LIMIT 50
  Fetch since offset:     SELECT WHERE chat_id = ? AND message_id > ? LIMIT 50

---

### 6. Delivery Status — sent → delivered → read

  Sent:
    Message accepted by Message Service and written to DB.
    Sender's WS Server acknowledges receipt → client shows single tick.

  Delivered:
    Recipient's WS Server successfully writes to recipient's WebSocket.
    WS Server sends "delivered" event back to sender via Redis pub/sub.
    Sender's client shows double tick.

  Read:
    When recipient scrolls to/opens the message, client sends a
    read receipt over the WebSocket connection (not a separate HTTP call).
    Avoids HTTP overhead for every message read.
    Server updates message status in DB → fan-out to sender → blue ticks.

  Group read receipts:
    Each member sends their own read receipt.
    Message shows read only when ALL members have sent read receipts.
    DB tracks per-member read status: (group_id, message_id, user_id) → read_at

---

### 7. WebSocket Server — What Lives Where

Common mistake: storing all user state on the WebSocket server.
This forces consistent hashing (same user always routes to same server).

Correct split:

  On WebSocket server (unavoidable — TCP socket is an OS resource):
    { user_id → open WebSocket socket object }

  In Redis (should NOT be on server):
    presence:{user_id}          → "online" with TTL 30s
    sequence:{user_id}          → last delivered sequence number
    group_members:{group_id}    → set of user_ids in the group

  Why this matters:
    If state is on the server → must always route user to same server
    → consistent hashing required → complex rebalancing on scale-out.

    If state is in Redis → any server can serve any user → no consistent
    hashing needed → load balancer uses least-connection freely.

Presence via heartbeat:
  Client sends heartbeat every 20s over WebSocket.
  Server: SETEX presence:{user_id} 30 "online"
  Heartbeats stop → TTL expires → user shown offline.

---

### 8. Redis pub/sub Channels

Direct message channel:
  "dm:{smaller_user_id}:{larger_user_id}"
  (sort user IDs so channel name is canonical regardless of who sends)

  Sender A (id=5) → Receiver B (id=42): channel = "dm:5:42"
  WS Server holding B's connection subscribes to "dm:5:42"
  Message Service publishes to "dm:5:42"

Group message channel:
  "group:{group_id}"
  All WS servers holding connections for group members subscribe.
  One PUBLISH → all member connections receive the message.

---

### 9. Sequence Numbers

Every message has a sequence number assigned by the server.
Monotonically increasing per chat (direct or group).

Purpose:
  1. Ordering: canonical order regardless of client clock skew
  2. Gap detection: client receives seq=1,2,4 → seq=3 missing → request replay
  3. Deduplication on reconnect: client tells server last received seq,
     server sends only seq > that value

Client reconnect flow:
  Client reconnects → sends {chat_id, last_seq: 47}
  Server fetches messages WHERE chat_id = ? AND seq > 47 FROM Cassandra
  Sends missed messages → delivery resumes seamlessly

---

### 10. Group Chat

Redis channel: "group:{group_id}"
All WS servers with group members subscribed receive every PUBLISH.
Each server writes only to connections of members in that group it holds.

Group membership in Redis:
  SMEMBERS group_members:{group_id} → set of user_ids

On new message:
  Kafka Consumer reads group message
  → PUBLISH "group:{group_id}" to Redis
  → All subscribed WS servers receive
  → Each server cross-references its local connections with group members
  → Delivers to members it holds connections for

Offline group members:
  WS server has no connection for them → push notification sent
  → they reconnect and fetch missed messages using sequence number

---

### 11. Media / File Handling

Large files cannot go through WebSocket — too slow, blocks the connection.

Upload flow:
  Client → POST /upload → Upload Service → S3 / object storage
  Upload Service returns: S3 URL + media_id
  Client sends message via WebSocket with {type: "media", media_url: S3_URL}
  Recipient downloads file directly from S3/CDN — not through WS server

Why CDN:
  Media is immutable once uploaded. CDN caches it at edge nodes.
  Recipient fetches from nearest CDN node — fast, offloads S3.

Thumbnail:
  Upload Service generates thumbnail on upload.
  Message carries thumbnail (small, safe to inline) + full media URL.
  Recipient sees thumbnail instantly, downloads full file on tap.

---

### 12. Kafka Partitioning

Direct chat: partition key = canonical chat ID (sorted user IDs)
  "chat:5:42" → same partition for all messages in this conversation
  Guarantees ordering within a conversation.

Group chat: partition key = group_id
  "group:99" → same partition for all messages in this group
  Guarantees ordering within a group.

Hot partition for popular groups (1,000 members, very active):
  A single partition handles ~100MB/s — sufficient for any group chat.
  1,000 members sending text messages is far below this limit.
  group_id alone as partition key is correct.

  DO NOT use group_id + hour — time-based suffixes do not distribute
  identity load. They create time buckets, not cardinality distribution.
  If a partition is hot, increase partition count and rekey by identity
  (e.g. group_id + shard_number assigned at group creation).

---

### 13. Redis Cluster — Sharding

Single Redis node cannot handle 2B users.
Redis Cluster shards channels across nodes by channel name hash.

  "dm:5:42"     → hash → shard 3
  "group:99"    → hash → shard 7
  "dm:100:200"  → hash → shard 1

Each WS server connects to the correct shard for each channel it subscribes to.
Redis Cluster handles routing transparently.

For extremely active group channels (1,000 members, high message rate):
  CPU bottleneck: one PUBLISH to a channel with many subscribers saturates
  the Redis node's event loop.
  Fix: shard the group channel — "group:99:shard:0..N"
  Publisher writes to all shards. Each shard covers a subset of WS servers.

---

### 14. L7 Load Balancer Configuration

WebSocket connections are long-lived. Two critical settings:

  Sticky sessions OR Redis state:
    If user state is in Redis → no sticky sessions needed (preferred)
    If user state is on server → sticky sessions via consistent hashing

  Long connection timeout:
    Default LB timeout kills WebSocket connections after 60s.
    Set to hours or disable for WebSocket endpoints.

  WebSocket-aware LB:
    Must understand HTTP upgrade handshake (101 Switching Protocols).
    Use wss:// (WebSocket over TLS) — passes through proxies that
    would otherwise block or corrupt the upgrade handshake.

---

## Scaling Numbers

  100B messages/day = ~1.15M messages/second peak
  Kafka: partition by chat_id, ~100 partitions handles this easily
  WS servers: 50k connections each → 2B users, 10% online = 200M connections
              → 4,000 WS servers
  Cassandra: ~1.15M writes/sec → 50-node cluster with replication factor 3
  Redis: cluster of 20-50 nodes depending on channel count and message rate

---

## Decision Tree (questions to ask in interview)

1. Bidirectional communication needed?
   Yes → WebSocket. Not SSE.

2. Where is user state stored?
   In Redis → stateless WS servers, any server, least-connection LB
   On server → consistent hashing required, harder to scale

3. When do you write to DB?
   Always BEFORE delivery attempt, not after. Persistence first.

4. How do offline users get messages?
   DB stores message + Push notification (APNs/FCM) wakes device
   → reconnect → fetch from DB using sequence number

5. How do you handle media?
   Separate upload service → S3 → CDN. Never through WebSocket.

6. Group chat channel fan-out bottleneck?
   High message rate → shard the Redis channel
   Too many connections → hierarchical fan-out

---

## Interview Signals (Senior Level)

Weak answers:
  "Store messages in DB when delivered."
  → Wrong. Store first, deliver second. Offline messages would be lost.

  "Use SSE for WhatsApp."
  → SSE is unidirectional. Client cannot send messages back on SSE.

  "Use consistent hashing to route users to the same WS server."
  → True, but missed: if state is in Redis, consistent hashing is unnecessary.
    Externalizing state is the better architecture.

  "Use group_id + hour as Kafka partition key for hot groups."
  → Time suffixes don't distribute identity load. Wrong fix.

Strong answers:
  "Messages are written to Cassandra before any delivery attempt.
  Redis fan-out is best-effort on top of the durable DB write."

  "WebSocket servers are stateless — presence, sequence numbers, and
  group memberships live in Redis. Any server can serve any user.
  No consistent hashing needed, LB uses least-connection freely."

  "Offline users receive a push notification via APNs/FCM that wakes
  the app. On reconnect, the client sends its last sequence number
  and the server replays missed messages from Cassandra."

---

## Staff / Principal Level — Additional Depth

### 15. Group Fan-out at Scale

Current design: Kafka consumer publishes to "group:99", Redis delivers
to every WS server holding a group member's connection.

Problem at scale:
  10,000 active groups × 1,000 members × 10 messages/sec
  = 100M Redis deliveries/sec
  Single Kafka consumer publishing inline cannot keep up.

Fix: dedicated Group Fan-out Service

```
Kafka (group message topic)
      ↓
Group Fan-out Service (horizontally scaled, separate from DM path)
      ↓
Redis PUBLISH "group:99:shard:0..N"
      ↓
WS Servers → deliver to connected group members
```

Why separate service:
  Group fan-out is O(members) per message. Direct message is O(1).
  Mixing them means a viral group message slows down all DM delivery.
  Separate services = separate scaling, separate failure domains.

---

### 16. Distributed Message ID — Snowflake over UUID

Document mentions UUID v7 or Snowflake. The choice matters for Cassandra.

UUID (random):
  Not time-ordered. Used as Cassandra clustering key → random write
  scatter across the partition → poor write performance, read amplification.

Snowflake ID (Twitter's approach):
  64-bit integer = 41-bit timestamp + 10-bit machine ID + 12-bit sequence
  Time-ordered → sequential writes into Cassandra partition → fast appends
  Supports 4,096 IDs per millisecond per machine
  Globally unique without coordination between machines

Why it matters:
  Cassandra LSM tree loves sequential writes. Random writes cause
  compaction overhead and slow range scans. Always use time-ordered IDs
  as clustering keys in Cassandra.

---

### 17. Multi-Device Sync

A user has WhatsApp on phone, laptop, and web simultaneously.
The document tracks sequence:{user_id} — one cursor per user. Wrong.

Phone at seq=47, laptop at seq=44 are different states.
Each device must track its own delivery cursor.

Correct Redis keys:
  sequence:{user_id}:{device_id} → last delivered seq per device

On reconnect:
  Device sends: {chat_id, device_id, last_seq: 44}
  Server fetches: WHERE chat_id = ? AND seq > 44
  Sends missed messages to THIS device only

Device registration:
  On first install, app registers a device_id with the server.
  Stored in a devices table: (user_id, device_id, platform, push_token)
  Push notifications sent per device (each device has its own APNs/FCM token).

---

### 18. Thundering Herd on WS Server Failure

A WS server holding 50,000 connections crashes.
All 50,000 users detect disconnection and reconnect within 3 seconds.
Each reconnection fetches missed messages from Cassandra.
50,000 concurrent range queries hit Cassandra simultaneously → overloaded.

Fix: exponential backoff with jitter on client reconnect.

```
reconnect_delay = min(base * 2^attempt, max_delay) + random_jitter

base = 1s, max = 60s, jitter = random(0, 1s)

attempt 1: ~1s + jitter
attempt 2: ~2s + jitter
attempt 3: ~4s + jitter
...
```

Jitter is critical — without it, all 50,000 clients retry at the same
intervals (1s, 2s, 4s...) and the thundering herd repeats on every backoff.
Jitter spreads reconnections over time, Cassandra absorbs the load gradually.

---

### 19. Conversation Inbox Table

The conversation list (recent chats with last message + unread count)
requires a separate data model — not derivable from the messages table
efficiently at read time.

```
inbox table (Cassandra):
  partition key:   user_id
  clustering key:  last_message_at DESC
  columns:         chat_id, last_message_preview, unread_count,
                   last_message_id, updated_at
```

Every new message triggers two writes:
  1. INSERT into messages table (the actual message)
  2. UPDATE inbox table for both sender and all recipients
     → update last_message_preview, increment unread_count

Why not compute at read time:
  Fetching the last message per conversation for 1,000 conversations
  requires 1,000 Cassandra reads → too slow for conversation list load.
  Pre-computed inbox table = single read for the full conversation list.

Unread count:
  Incremented on message receipt.
  Reset to 0 when user opens the conversation (read receipt sent).
  Stored as a counter in the inbox table.

---

### 20. Geo-Distribution

2B users globally. Sender in India, recipient in USA.

Regional architecture:
  - GeoDNS routes users to nearest WS server cluster
  - Each region has its own: WS servers, Kafka cluster, Cassandra cluster
  - Users connect to regional cluster — low latency for local messages

Cross-region message delivery:
```
Sender (India) → India WS Server → India Kafka
                                        ↓
                              Kafka MirrorMaker (replication)
                                        ↓
                                  USA Kafka
                                        ↓
                              USA Kafka Consumer → USA Redis
                                        ↓
                              USA WS Server → Recipient (USA)
```

Latency: India → USA cross-region adds ~150ms but stays within 2s SLA.
Local messages (same region): ~50ms end-to-end.

Cassandra cross-region:
  Cassandra supports multi-datacenter replication natively.
  Replication factor 3 per region. Writes acknowledged locally,
  replicated asynchronously to other regions.
  Recipient in USA can read message from USA Cassandra node — no cross-region read.

---

### 21. End-to-End Delivery Guarantee and Dead Letter Queue

Current design delivers at-least-once (Kafka) + client deduplicates
by message_id. But what detects a stuck message?

Failure scenario:
  Kafka consumer publishes to Redis → WS server delivers to client
  → WS server crashes before sending "delivered" receipt back
  → Kafka consumer never receives confirmation
  → Message is in Cassandra but status stuck at "sent" forever

Fix — timeout-based retry with DLQ:

  1. Message Service writes message with status=SENT and delivered_at=NULL
  2. A background job queries:
       SELECT * FROM messages WHERE status=SENT AND created_at < NOW() - 30s
  3. For each stuck message: re-trigger delivery attempt
  4. After N retries with no delivery: move to Dead Letter Queue (DLQ)
  5. DLQ triggers alert to on-call — manual investigation

Monitoring:
  Track p99 latency from SENT → DELIVERED per region.
  Alert if p99 > 5 seconds — indicates delivery pipeline is degraded.
  Dashboard showing DLQ depth — non-zero = active delivery failures.

---

### 22. End-to-End Encryption — Architecture Implications

WhatsApp uses the Signal protocol. Server stores encrypted blobs only.
This is not just a security feature — it shapes the data model.

What the server CANNOT do because of E2E encryption:
  - Generate last message preview for inbox table (content is encrypted)
  - Search message content
  - Moderate message content server-side

What changes in the architecture:
  - Inbox table last_message_preview: stores encrypted preview, rendered
    only by the recipient's device using their local key
  - Media: client encrypts file before upload → server stores encrypted blob
    → encryption key is sent inside the (encrypted) message payload
  - Key exchange: requires a Key Distribution Service (KDS) that stores
    each device's public key. Before first message, sender fetches
    recipient's public key from KDS → establishes session key → encrypts

Key Distribution Service:
  Stores: (user_id, device_id) → public_key
  Read-heavy, rarely written (only on device registration or key rotation)
  Cache aggressively in Redis — key lookups happen before every new conversation

---

## Knowledge Gaps Identified During Study Session

Gap 1 — SSE vs WebSocket for chat:
  SSE is unidirectional. Chat requires bidirectional. Always WebSocket for chat.

Gap 2 — DB write order:
  Must persist to DB BEFORE delivery attempt.
  Never store only on successful delivery — offline messages lost.

Gap 3 — Push notifications for offline users:
  Storing in a queue is not enough. Device must be woken via APNs/FCM.
  Without push, offline user never knows a message arrived.

Gap 4 — State on server vs state in Redis:
  State on server → consistent hashing required.
  State in Redis → stateless servers, no consistent hashing, easier scaling.
  Only the TCP socket object must live on the server.

Gap 5 — Media handling:
  Large files never go through WebSocket.
  Separate upload service → S3 → CDN URL sent in message.

Gap 6 — Group chat partition key:
  group_id alone is correct. Time-based suffixes don't distribute load.
  A single Kafka partition handles ~100MB/s — more than enough for any group.

Gap 7 — Kafka Consumer as bridge:
  Kafka does not talk to Redis directly.
  A Kafka consumer reads events and calls Redis PUBLISH.
  This component must be explicitly named in the architecture.
