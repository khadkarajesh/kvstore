## Back-of-Envelope Estimation

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

### Why Estimation Matters in Interviews

  Interviewers test: can you reason at scale? do you know orders of magnitude?
  Expected in first 5 minutes of every FAANG system design interview.

  What it signals:
  - You understand the difference between 10K QPS and 10M QPS (those require very different architectures)
  - You can drive the conversation: your numbers justify your design choices
  - You are not guessing blindly when you say "we need a cache" or "we need sharding"

  Common failure mode: jumping to design without anchoring on scale. Interviewers notice.

---

### Reference Numbers — Memorize These

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  LATENCY
  ─────────────────────────────────────────────
  L1 cache hit                  0.5 ns
  L2 cache hit                  7 ns
  RAM access                    100 ns
  SSD random read               150 µs
  Network RTT same datacenter   500 µs
  SSD sequential read (1MB)     1 ms
  HDD seek                      10 ms
  Network RTT cross-region      150 ms

  Key ratios to remember:
  - RAM is 200× faster than SSD random read
  - SSD random read is 20× faster than HDD seek
  - Cross-region network is 300× slower than same-datacenter network

  STORAGE UNITS
  ─────────────────────────────────────────────
  KB   10^3   (1,000 bytes)
  MB   10^6   (1,000,000 bytes)
  GB   10^9
  TB   10^12
  PB   10^15

  TIME UNITS
  ─────────────────────────────────────────────
  1 day    = 86,400 seconds  ≈ 100K seconds
  1 month  ≈ 2.5M seconds
  1 year   ≈ 30M seconds

  COMMON QPS BENCHMARKS
  ─────────────────────────────────────────────
  Twitter reads     300K/sec
  Google search     100K/sec

  AVERAGE OBJECT SIZES
  ─────────────────────────────────────────────
  Tweet             280 chars ≈ 280 bytes
  Photo             ≈ 200 KB
  Video (1 minute)  ≈ 50 MB
  Webpage           ≈ 100 KB

---

### QPS Estimation

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Formula:
    QPS = DAU × actions_per_day / seconds_per_day
    Peak QPS = average QPS × 2–3×

  Worked example — Twitter reads:
    300M DAU × 10 reads/day / 100K sec = 30K QPS reads (average)
    Peak = 30K × 3 = 90K QPS

  Worked example — Twitter writes:
    300M DAU × 1 tweet/day / 100K sec = 3K QPS writes (average)
    → reads dominate by 10×, which justifies read-heavy caching strategy

  Interview move: always state the read:write ratio after computing both.
  That ratio directly motivates your caching, replication, and sharding choices.

---

### Storage Estimation

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Formula:
    storage_per_day = writes_per_day × avg_object_size
    total_storage   = storage_per_day × retention_days

  Worked example — Twitter:
    500M tweets/day × 280 bytes = 140 GB/day raw text
    5 years → 140 GB × 365 × 5 ≈ 255 TB
    With 3× replication → 255 TB × 3 ≈ 765 TB ≈ 1 PB

  Worked example — Twitter photos (assume 10% of tweets include photo):
    50M photos/day × 200 KB = 10 TB/day
    5 years with replication → 10 TB × 365 × 5 × 3 ≈ 55 PB

  Interview move: always apply the replication multiplier (3× is standard).
  That number justifies blob storage (S3, GCS) vs a relational DB.

---

### Bandwidth Estimation

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Formula:
    bandwidth = QPS × avg_object_size

  Ingress vs egress:
  - Ingress → upload bandwidth (what users send to you)
  - Egress  → download bandwidth (what you send to users)
  - Egress usually dominates for read-heavy systems

  Worked example — YouTube ingress:
    500 hours of video uploaded per minute
    → 500 × 60 min/video × 50 MB/min = 1.5 TB/min ingress
    → 1.5 TB / 60 sec ≈ 25 GB/sec ingress bandwidth

  Worked example — Twitter egress:
    30K QPS reads × 280 bytes/tweet ≈ 8 MB/sec
    → negligible for text; photos change this dramatically (30K × 200KB = 6 GB/sec)

---

### Memory / Cache Estimation

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Formula:
    cache_size = hot_data_fraction × total_active_data

  80/20 rule: 20% of data serves 80% of requests — cache that 20%.

  Worked example — Twitter timeline cache:
    140 GB/day tweets × 20% hot = 28 GB cache needed per day of hot data
    A single Redis node (128GB RAM) can hold several days of hot tweets

  Interview move: after computing cache size, state whether it fits on one node
  or requires a distributed cache (Redis Cluster). That choice is architectural.

---

### How to Present Estimation in an Interview

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Step-by-step delivery:
  1. State assumptions out loud before calculating
     → "I'll assume 300M DAU, 10 reads/day each, average tweet is 280 bytes"
  2. Round aggressively — directionally right beats precisely wrong
     → "86,400 ≈ 100K — I'll use 100K for easy math"
  3. Use powers of 10, not long multiplication
     → 300M × 10 / 100K = 3M / 100K = 30K
  4. One sentence conclusion that connects numbers to design
     → "30K QPS reads and 3K QPS writes means read:write = 10:1 — we want aggressive caching and can get away with fewer write replicas"

  What not to do:
  - Do not spend 10 minutes on math. 2–3 minutes max.
  - Do not skip the conclusion. Numbers alone mean nothing without design implication.
  - Do not make up numbers without stating they are assumptions.
