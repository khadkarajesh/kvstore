# API Design — Senior System Design Interview Notes

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
## 1. REST
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

PRINCIPLES
  → stateless: each request contains all information needed; server holds no client session state
  → resource-based URLs: /users/42, /orders/99/items (nouns, not verbs)
  → HTTP verbs carry semantics:
      GET    → read, safe, idempotent
      POST   → create, not idempotent
      PUT    → replace entire resource, idempotent
      PATCH  → partial update, not always idempotent
      DELETE → remove, idempotent
  → human readable: JSON over HTTP, debuggable with curl

WHEN TO USE
  → public-facing APIs (third-party developers, browser clients)
  → when human readability and wide tooling support matter
  → standard CRUD operations

WEAKNESSES
  → over-fetching: endpoint returns more fields than client needs
  → under-fetching: client needs data from multiple endpoints → multiple round trips
  → no native streaming support
  → versioning required when schema changes

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
## 2. gRPC
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

FUNDAMENTALS
  → built on HTTP/2: multiplexing, header compression, binary framing
  → uses Protocol Buffers: binary serialization, strongly typed, smaller payload than JSON
  → schema-first: .proto file defines contract; code generated for both sides
  → supports four communication patterns:
      Unary:                  client sends one request, receives one response
      Server streaming:       client sends one request, server streams many responses
      Client streaming:       client streams many requests, server responds once
      Bidirectional streaming: both sides stream simultaneously

WHEN TO USE
  → internal microservice-to-microservice communication
  → high throughput or latency-sensitive paths
  → streaming required (real-time updates, large data transfer)
  → polyglot environments (proto generates clients for Go, Java, Python, etc.)

WEAKNESSES
  → not human readable (binary): harder to debug without tooling
  → browser support limited (gRPC-Web is a workaround with a proxy)
  → harder to test without proto-aware tools (grpcurl helps)
  → requires schema file to be shared/versioned

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
## 3. GRAPHQL
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

FUNDAMENTALS
  → client specifies exactly what fields it needs in the query
  → single endpoint: POST /graphql
  → eliminates over-fetching and under-fetching from REST
  → schema defines types and relationships on the server

WHEN TO USE
  → complex data graphs with many relationships (social network: user → posts → comments → likes)
  → mobile clients constrained by bandwidth (request only the fields you need)
  → rapid frontend iteration (frontend changes query, no backend API change needed)
  → multiple clients needing different shapes of the same data

WEAKNESSES
  → N+1 QUERY PROBLEM: fetching a list of users, then for each user fetching their posts → N+1 DB queries
      fix: DataLoader (batching + caching resolver calls)
  → caching harder: all requests are POST, not GET → HTTP caching not automatic
  → server-side complexity: resolvers for every field, schema maintenance
  → not suitable for simple CRUD with no complex relationships

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
## 4. THE DECISION — WHEN TO USE WHICH
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  External public API                    → REST
  Internal microservice communication    → gRPC
  Complex relational data, mobile clients → GraphQL
  Real-time bidirectional communication  → WebSockets
  Event streaming, pub/sub              → Kafka / message queue (not REST/gRPC)

INTERVIEW SCRIPT
  "For the client-facing API I would use REST with cursor-based pagination.
   For internal service communication I would use gRPC for performance and type safety.
   Every mutation endpoint requires an idempotency key to handle client retries safely."

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
## 5. PAGINATION STRATEGIES
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

OFFSET PAGINATION
  → query: SELECT * FROM posts ORDER BY id LIMIT 10 OFFSET 100
  → client sends: ?page=11&per_page=10
  → problems:
      UNSTABLE: new row inserted at offset 50 while user is on page 5 → page 6 skips or duplicates a row
      SLOW AT SCALE: DB must scan and discard all rows before offset → O(offset) cost
  → use only for: small datasets, admin panels, when simplicity matters more than correctness

CURSOR-BASED PAGINATION (preferred at scale)
  → client sends: ?after=cursor_xyz&limit=10
  → cursor encodes position in the result set (opaque to client: often base64 of last row's id or timestamp)
  → server query: SELECT * FROM posts WHERE id > :cursor ORDER BY id LIMIT 10
  → advantages:
      STABLE: new inserts don't shift pages — cursor points to a specific row, not a count
      FAST: DB uses index to jump directly to cursor position → O(log N) regardless of offset
  → used by: Twitter, Facebook Graph API, Stripe, GitHub API
  → rule: use cursor-based for any paginated list that changes frequently or has large offsets

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
## 6. API VERSIONING
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

URL VERSIONING
  → /api/v1/users, /api/v2/users
  → visible in URL: clear, discoverable, easy to test in browser, easily cached
  → URL changes on version bump (some consider this REST violation)
  → use for: public APIs (consumers need clear version visibility)

HEADER VERSIONING
  → Accept: application/vnd.myapi+json;version=2
  → URL stays clean, no breaking change to URL structure
  → harder to test (cannot just paste URL in browser)
  → use for: internal APIs where consumers control request headers

WHICH TO RECOMMEND IN INTERVIEW
  → public API → URL versioning (/v1/, /v2/)
  → internal API → header versioning or gRPC (which versions via .proto field numbers)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
## 7. IDEMPOTENCY KEYS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

THE PROBLEM
  → client sends payment request → network timeout → did the request succeed?
  → if client retries: risk of double charge
  → cannot safely retry without idempotency guarantee

THE SOLUTION
  → client generates a unique idempotency_key (UUID) before sending request
  → client sends: POST /payments { idempotency_key: "uuid-abc", amount: 100 }
  → server stores: idempotency_key → response (in Redis or DB)
  → on first request: execute payment, store result under key
  → on duplicate request (same key): return stored result, do NOT execute again

IMPLEMENTATION
  → store key in Redis with TTL (e.g. 24 hours)
  → on duplicate key within TTL: return cached response with 200 (not 4xx)
  → after TTL expires: key gone, treat as new request (client should not retry after TTL)
  → use distributed lock on key to handle concurrent duplicates

RULE
  → every mutation with side effects needs idempotency key:
      payments, order placement, email sends, notification dispatches, inventory decrements

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
## 8. WHAT TO SAY IN A SENIOR INTERVIEW
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

After defining any API:

  "For the client-facing API I would use REST with cursor-based pagination.
   For internal service communication I would use gRPC for performance.
   Every mutation endpoint requires an idempotency key stored in Redis
   to handle client retries safely without duplicate side effects."

If asked about pagination specifically:

  "I would use cursor-based pagination rather than offset. Offset pagination
   is unstable under concurrent writes and degrades to O(offset) scans at scale.
   Cursor-based uses an indexed seek and is stable regardless of concurrent inserts."
