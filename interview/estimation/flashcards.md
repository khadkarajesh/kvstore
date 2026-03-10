## Estimation — Flashcards

Format: Q → A
Cover the answer, recall it, then check.

---

**Q: What is the latency of an L1 cache hit?**
A: 0.5 ns

---

**Q: What is the latency of an L2 cache hit?**
A: 7 ns

---

**Q: What is the latency of a RAM access?**
A: 100 ns

---

**Q: What is the latency of an SSD random read?**
A: 150 µs (microseconds)

---

**Q: What is the network round-trip time within the same datacenter?**
A: 500 µs (microseconds)

---

**Q: What is the latency of an SSD sequential read for 1MB?**
A: 1 ms

---

**Q: What is the latency of an HDD seek?**
A: 10 ms

---

**Q: What is the network round-trip time cross-region?**
A: 150 ms

---

**Q: How many seconds are in a day? Month? Year?**
A: Day ≈ 100K (86,400). Month ≈ 2.5M. Year ≈ 30M.

---

**Q: What is the formula for average QPS?**
A: QPS = DAU × actions_per_day / seconds_per_day (use 100K for seconds_per_day)

---

**Q: What multiplier is used to estimate peak QPS from average QPS?**
A: 2–3×

---

**Q: What is the formula for daily storage?**
A: storage_per_day = writes_per_day × avg_object_size

---

**Q: What replication multiplier should you always apply to storage estimates?**
A: 3× (standard replication factor)

---

**Q: What is the 80/20 rule as applied to caching?**
A: 20% of data serves 80% of requests — cache that 20% to absorb most of the read load.

---

**Q: What is the formula for cache size?**
A: cache_size = hot_data_fraction × total_active_data. Use 20% as the default hot fraction.

---

**Q: What is the average size of a tweet?**
A: 280 chars ≈ 280 bytes

---

**Q: What is the average size of a photo?**
A: ≈ 200 KB

---

**Q: What is the average size of 1 minute of video?**
A: ≈ 50 MB

---

**Q: What is the 4-step process for presenting estimation in an interview?**
A: 1) State assumptions out loud. 2) Round aggressively. 3) Use powers of 10. 4) One sentence conclusion connecting numbers to a design decision.

---

**Q: Twitter has 300M DAU, each user reads 10 tweets/day. What is average read QPS? Peak?**
A: Average = 300M × 10 / 100K = 30K QPS. Peak = 30K × 3 = 90K QPS.
