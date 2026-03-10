# 3-Month System Design Interview Preparation Plan

Goal: Build genuine intuition for system design interviews through paper reading,
spaced repetition, and deliberate practice.

---

## The Core Principle

Reading papers alone does not build interview intuition. Applying the thinking
to new problems does. The paper gives you raw material — what you do after
reading determines whether it sticks.

The loop that builds intuition:

  Read → Notes → Flashcards → Answer questions out loud → Apply to new problem
    ↑                                                              │
    └──────────────── spaced repetition every week ───────────────┘

---

## Why Most Plans Fail (and How This One Avoids It)

The common pattern: excited at start → a few days pass → feels like no progress
→ drop the plan → anxious to return → avoid returning.

This plan addresses each stage directly:

  "No progress" feeling:
  → Progress is measured weekly with a binary test, not monthly checkpoints.
    Every Sunday you close your notes and write for 5 minutes. If you can
    write it fluently, you own it. No self-convincing possible.

  Rigid scheduling breaks on first missed day:
  → Plan uses activity quotas per week, not day-of-week assignments.
    Miss Tuesday — do it Wednesday. The week still succeeds.

  Weekend sessions feel like work with no clear end:
  → Every design session has a template with four required outputs.
    When all four exist in writing, you are done. No ambiguity.

  Weeks 3-5 energy dip:
  → Quota is explicitly reduced in dip weeks by design. Not failure —
    built into the plan. Reduced load prevents the dip from feeling
    like falling behind.

  No accountability:
  → One external checkpoint per month. One person hears your explanation.
    Knowing someone will hear it creates a forcing function that
    self-review cannot replicate.

  Anxiety about returning after a gap:
  → Re-entry ritual is defined. Follow it exactly, no exceptions.

---

## Minimum Viable Day

On any day — busy, tired, unmotivated — this is the floor:

  Do 5 flashcards. That is it. 5 minutes.

  This counts as a full session. The streak is intact.
  The goal is to never break contact with the material entirely.
  5 cards beats zero cards every time.

  Never skip twice in a row. That is the only rule.

---

## Re-Entry Ritual (after missing 3+ days)

Do not try to catch up. Do not feel guilty. Follow these steps exactly:

  1. Open weak-points.md. Read it. Do not add anything yet.
  2. Do 10 flashcards from whichever paper you were last on.
  3. Write one sentence: "I am back. Today I reviewed [topic]."
     Save it in weak-points.md under a new date entry.
  4. Resume the plan at the current week. Do not go back.

Why: The anxiety of returning is always worse than the actual return.
     The ritual removes the decision of "where do I start" — you always
     know exactly what to do first. That removes the friction.

---

## Weekly Progress Test (Binary)

Every Sunday — close notes, open a blank document, write for 5 minutes:

  Explain one concept from this week as if teaching someone who
  knows databases but has never read this paper.

  If you write fluently without stopping: you own it. Mark it done.
  If you stall after 2 sentences: it goes on weak-points.md as a gap.
    Do not mark it done. Revisit next week.

  This is binary. No partial credit. No "I kind of understood it."
  The test removes self-convincing — the gap between recognition
  and recall is exactly where interview performance collapses.

Also record these three questions in weak-points.md every Sunday:
  1. Which topic was hardest this week?
  2. Which concept felt most natural?
  3. One thing I can explain now that I could not last week.

Question 3 makes progress visible on the days it feels invisible.

---

## External Accountability (Monthly)

End of each month: find one person and explain one system design to them.

  Not a mock interview. Not a formal session.
  Just: "Let me explain how Bigtable works and when you would use it."
  15-20 minutes. Any person — friend, colleague, study partner.

  Why this works: knowing someone will hear your explanation in 4 weeks
  changes how you study this week. Self-review has no consequence.
  External explanation does.

  If you have no one: record yourself on your phone explaining it.
  Play it back. You will immediately hear what you do not understand.
  It is uncomfortable. That discomfort is the point.

---

## Design Session Template

Every design session produces exactly these four outputs in writing.
When all four exist, the session is done. Not before.

  Output 1 — Estimation (before any diagram):
    QPS, storage, bandwidth calculated with reference numbers.
    One sentence conclusion linking numbers to architecture decisions.

  Output 2 — Requirements + constraints (1 paragraph):
    Read/write patterns, scale, consistency, availability target.
    At least one constraint stated explicitly ("this means I cannot use X").

  Output 3 — Technology choice with ruled-out alternatives (1 paragraph):
    "I chose X because Y. I ruled out A because B. I ruled out C because D."
    Must name at least two ruled-out alternatives with reasons.

  Output 4 — Diagram (ASCII is fine):
    Components, data flow, where data lives. Include CDN/LB/cache/queue.
    Must exist. Thinking without drawing is not a design.

  Output 5 — Failure scenario (1 paragraph):
    What is the hardest failure case? How does the system handle it?
    What is lost or degraded during that failure?

  Output 6 — Senior signal (1 paragraph — this is what separates senior):
    THREE things in one paragraph:
    1. Capacity: "At X QPS I need approximately N servers / M Redis nodes"
    2. Observability: RED metrics, distributed tracing, health checks, alerts
    3. Evolution: "V1 bottleneck is X. V2 would change Y to handle 10x scale"

  Total time: 40-50 minutes. If step 6 is missing, redo it — do not skip.
  If it takes less than 30 minutes, you skipped something.

---

## Deliberate Practice System

Deliberate practice is not doing more of the same. It is specifically
targeting the point where you fail and drilling that point with increasing
difficulty until it stops being a failure point.

High performers do not practice what they are good at.
They find exactly where they break and work there.

---

### The Four Drills

Each drill is short (5-15 min), targeted, and produces immediate feedback.
Do 2 drills per week. Always pick drills that attack something in weak-points.md.

---

#### Drill 1: Constraint Removal (10 min)

Take a system you already designed. Remove one component. Re-solve it.

Why: Forces you to understand WHY each component exists, not just WHAT it does.
     If you cannot re-solve without it, you did not understand it — you memorised it.

How:
  1. Pick a system you designed this week.
  2. Remove one component (e.g. remove the cache, remove Kafka, remove the leader).
  3. Write: what breaks? What do you use instead? What do you give up?
  4. Compare to your original design. Write the tradeoff explicitly.

Example prompts:
  - Design Gmail without Bigtable. What do you use? What do you give up?
  - Design the web crawl table without reversed row keys. What breaks?
  - Design Twitter's feed without any cache layer. What do you do instead?
  - Design a metrics system without Kafka. How do you handle ingestion spikes?
  - Design a consistent KV store without a leader. What problem does that create?

---

#### Drill 2: Scale Explosion (10 min)

Take a working design. 10x the scale. What breaks first? Then 100x.

Why: Forces you to understand the limits of every choice you made.
     Most candidates design for the happy path. Interviewers probe the limits.

How:
  1. Take any design (yours or from HelloInterview).
  2. Write: at 10x scale, what is the first bottleneck?
  3. Fix that bottleneck.
  4. Write: at 100x scale, what breaks in your fix?

Example prompts:
  - Your Gmail design handles 1M users. At 100M users, what breaks first?
  - Your Bigtable tablet is 1GB. Now tablets are 1TB each. What changes?
  - Your Dynamo cluster has 10 nodes. Now it has 10,000. What new problems appear?
  - Your metrics system ingests 10K events/sec. Now it is 10M/sec. What breaks?
  - Your consistent KV store has 3 nodes. Now it has 300. What happens to latency?

---

#### Drill 3: Adversarial (10 min)

Make the strongest possible argument against your own design. Then defend it.

Why: Forces you to know your tradeoffs deeply enough to argue both sides.
     Interviewers attack your design. If you cannot attack it yourself,
     you will freeze when they do.

How:
  1. Take a design you produced this week.
  2. Spend 5 minutes writing the best case AGAINST your technology choice.
     Be specific. Use the decision framework against yourself.
  3. Spend 5 minutes defending your original choice.
     Acknowledge the attack. Explain why you still chose it anyway.

Example attack prompts:
  - You chose Bigtable. Attack: "Why not Cassandra? It also has wide columns
    and gives you higher availability. Defend your choice."
  - You chose Dynamo. Attack: "Why not Spanner? You could have transactions
    and still scale horizontally. What are you giving up and is it worth it?"
  - You chose Kafka. Attack: "Why not just write to a database and poll it?
    You are adding operational complexity for what exactly?"
  - You chose CP. Attack: "Your users will see outages. Is that really better
    than briefly stale data? Defend your consistency requirement."
  - You designed for range scans. Attack: "Most queries are single key lookups.
    You over-engineered this. Argue why range scans justify the complexity."

---

#### Drill 4: Speed Comparison (5 min)

90 seconds. Compare two technologies for a specific use case. Write it.

Why: Interviews test pattern recognition under time pressure.
     You need to produce a clear comparison without thinking from scratch.
     Speed drills build the muscle memory for instant recall.

How:
  1. Set a 90-second timer.
  2. Write your comparison before the timer ends.
  3. After: check your notes. What did you miss? What was wrong?
  4. Add misses to weak-points.md.

Example prompts (90 seconds each):
  - Bigtable vs Cassandra for a web crawl index. Go.
  - Dynamo vs Spanner for a bank ledger. Go.
  - Kafka vs a database poll for real-time notifications. Go.
  - Bigtable vs PostgreSQL for storing 1 billion user events. Go.
  - Cassandra vs Redis for a session store with 100M active users. Go.
  - Dynamo vs Cassandra for a shopping cart. Go.
  - Spanner vs Bigtable for a globally consistent inventory system. Go.
  - Kafka vs Dynamo for an activity feed. Go.
  - Elasticsearch vs PostgreSQL full-text search for product search. Go.
  - JWT vs session cookies for stateless API auth. Go.
  - WebSockets vs SSE vs long polling for a live feed. Go.
  - Object storage vs block storage for user-uploaded images. Go.
  - Active-active vs active-passive for a global e-commerce site. Go.
  - Token bucket vs sliding window for API rate limiting. Go.
  - 2PC vs Saga for a cross-service payment. Go.

---

### Deliberate Practice Weekly Schedule

Minimum: 2 drills per week. Pick based on weak-points.md.

  If weak-points.md shows gaps in technology comparison → do Speed drills
  If weak-points.md shows gaps in failure handling → do Constraint drills
  If weak-points.md shows gaps in scale reasoning → do Scale drills
  If weak-points.md shows you freeze when challenged → do Adversarial drills

Always pick the drill that attacks your current weakest area.
Never pick the drill you find easiest. That is practice, not deliberate practice.

---

### Motivating Yourself Through Deliberate Practice

Deliberate practice is uncomfortable by design. You will fail drills.
That is not a problem — it is the mechanism. Failure during practice
means you found a gap before the interview found it for you.

Track your drill results in weak-points.md with a simple marker:
  [date] | DRILL | [drill type] | [what broke] | [what you learned]

After 4 weeks, look back. You will see concepts you failed consistently
that you can now do fluently. That visible arc is what motivation feeds on.
It is not about feeling ready — it is about seeing that you are getting better.

---

## How to Read a Paper (Repeatable Process)

Use this process for every paper. Do not skip steps.

  Step 1: Before reading — draft answers from intuition
          Write what you think the paper will say. This surfaces gaps
          before you read, so you know what to look for.

  Step 2: Read in the given section order
          Do not read top to bottom. Read the sections specified in
          the reading guide. Skip what is marked skip.

  Step 3: For each section — write notes in your own words
          Do not copy sentences from the paper. Paraphrase.
          If you cannot paraphrase it, you do not understand it yet.

  Step 4: Compare your draft (Step 1) to your notes (Step 3)
          The delta is your learning. Write it down explicitly.

  Step 5: Build flashcards from your gaps, not from what you already knew

  Step 6: Answer the retention questions without looking at notes

Why: Passive reading produces recognition, not recall. Interviews test recall.

---

## Paper Reading Order and Why

  Already done:
  - Dynamo (AP KV store) — baseline for availability-first thinking
  - Bigtable (CP wide-column store) — baseline for consistency-first thinking

  Month 1: Spanner
  - Why now: Spanner is Bigtable + global transactions + SQL. Reading it
    forces you to articulate exactly what Bigtable cannot do and why.
    The TrueTime mechanism is the most asked-about distributed systems
    concept in Google interviews. Builds directly on Bigtable knowledge.
    Reading it immediately after Bigtable maximises retention overlap.

  Month 2: Kafka
  - Why: Every system design with "event stream", "activity feed", or
    "real-time analytics" needs Kafka. It introduces the log as a
    universal data structure. Cannot be substituted by any storage paper.

  Month 2: Cassandra
  - Why: Cassandra is Dynamo's replication + Bigtable's data model.
    Reading it synthesizes both papers. After Dynamo and Bigtable
    separately, Cassandra will take half the time and double the retention.

  Month 3: No new papers
  - Why: New material at interview stage creates surface knowledge that
    breaks under pressure. Month 3 is synthesis and simulation only.

---

## System Design Answer Framework (Senior Level)

Use this structure for every design session and mock interview.
Steps 1-6 produce a mid-level answer. Step 7 is what makes it senior.

  1. Clarify requirements (2 min)
     - What are the read and write patterns?
     - What scale? (users, requests/sec, data size)
     - What consistency is required?
     - What availability target? (99.9%? 99.99%?)
     - Real-time or eventual delivery?

  2. Estimation — do this before any diagram (2 min)
     - Calculate QPS, storage, bandwidth using reference numbers
     - State peak QPS = average × 3
     - One sentence conclusion: "We need ~X QPS and ~Y TB storage,
       so single machine is out and we need Z approach"
     - Interviewers who see no estimation before design mark it as mid-level

  3. State the constraints out loud (1 min)
     - "This is read-heavy so I will optimize for reads"
     - "Consistency is critical so I will choose CP"
     - "99.99% availability means no single points of failure"

  4. Rule out technologies using the decision framework (2 min)
     - Go through the 7 questions: Scale, Data shape, Access pattern,
       Write pattern, CAP, Query complexity, Latency
     - Name at least two alternatives and why you ruled them out

  5. High-level design (5 min)
     - Components and responsibilities
     - Data flow diagram
     - Include: CDN, load balancer, cache, storage, message queue
       where applicable — do not omit infrastructure layers

  6. Deep dive on the hardest part (10 min)
     - Row key / partitioning strategy
     - Failure handling — what breaks and how does system recover?
     - Consistency mechanism
     - API design — REST vs gRPC, pagination strategy, idempotency keys

  7. Senior signal — state all three proactively (2 min)
     Never wait to be asked. Raise these yourself.

     Capacity planning:
       "At 30K QPS with 1KB average response I need approximately
        X servers behind the load balancer"

     Operational concerns:
       "In production I would instrument with RED metrics on every
        endpoint, distributed trace IDs across all RPC calls,
        liveness + readiness health checks, and alert on p99
        latency and error rate — not CPU"

     Design evolution:
       "This is V1. At 10x scale the first bottleneck is X.
        V2 would address that by Y"

  8. State tradeoffs explicitly (1 min)
     - "I chose X which means I give up Y"
     - Never present a design without acknowledging its cost

  Definition of done for a design session:
  All 8 steps completed in writing. Step 7 must be present.
  If step 7 is missing, the answer is mid-level regardless of content.

Why: Senior interviewers score whether you think like someone who has
     operated production systems. Step 7 is that signal. Mid-level
     candidates wait to be asked about monitoring and evolution.
     Senior candidates raise it themselves.

---

## Mock Interview Debrief Template

After every HelloInterview session or mock interview, spend 10 minutes
writing this debrief. Without structure, post-interview analysis is vague
and learning does not accumulate.

  Date: ___________
  Problem: ___________
  Time taken: ___________

  Step-by-step self-score (1 = missed, 2 = partial, 3 = strong):

    [ ] Estimation done before design          ___
    [ ] Requirements clarified                 ___
    [ ] 2+ technologies ruled out with reasons ___
    [ ] Diagram drawn                          ___
    [ ] Failure scenario addressed             ___
    [ ] Senior signal stated (all 3)           ___
    [ ] Tradeoffs acknowledged                 ___
    [ ] Drove conversation (did not wait)      ___

  Total: ___ / 24

  What I could not answer:
  _________________________________

  What I said that was wrong:
  _________________________________

  One thing to add to weak-points.md:
  _________________________________

  Score trend: track total score across sessions.
  If score is not increasing after 3 sessions → that week's quota
  must include 2 extra deliberate practice drills on the lowest-scored step.

---

## HelloInterview Integration

You have a paid subscription. Use it as follows:

  Month 1 (Weeks 3-4):
  - After each design session, use HelloInterview to attempt the same
    problem in a structured format. Compare your written design to
    the model answer. Identify gaps.
  - Do not use HelloInterview before your own attempt. Always design
    first, then check. Using it as a reference first defeats the purpose.

  Month 2 (Weeks 5-8):
  - Replace one weekend design session per week with a HelloInterview
    timed session. This introduces time pressure earlier.
  - Use the HelloInterview problem bank to find problems that match
    the paper you just read that week.

  Month 3 (Weeks 9-12):
  - Use HelloInterview for 1-2 mock sessions per week, timed,
    with built-in feedback. Primary simulation tool in this phase.
  - After each session: add gaps directly to weak-points.md.

  Rule: HelloInterview is for simulation and gap identification.
        Your notes and flashcards are for retention.
        Do not use HelloInterview as a study tool — use it as a test.

---

## Month 1 — Foundation + Estimation + Caching

  Priority this month: before touching any new paper, build the two
  foundations that every interview needs — estimation and caching.
  Without these, Spanner knowledge does not help you pass interviews.

  Weeks 1-2 quota (full energy):
    3 flashcard sessions + 2 written question answers + 1 design session
    + 2 deliberate practice drills

  Weeks 3-4 quota (reduced — dip week by design):
    2 flashcard sessions + 1 written question answer + 1 design session
    + 2 deliberate practice drills (drills stay — they are short and targeted)

  Week 1: Estimation + Caching (new foundations)

    Read: interview/estimation/notes.md — full read, memorize reference numbers
    Read: interview/caching/notes.md — full read
    Flashcard sessions: 10 estimation cards + 10 caching cards
    Written questions: pick 3 from caching/interview-questions.md (no notes)
    Design session: "Design a URL shortener"
      Output 1: estimate QPS and storage before drawing anything
      Output 2: where does caching fit? which strategy?
      Output 3: diagram
      Output 4: what happens when cache layer goes down?
    Deliberate drills: 2 × Scale Explosion on any design you know

    Sunday test: given "100M DAU, each user makes 10 reads/day" —
    calculate QPS, storage for 1 year, cache size needed. No notes.

  Week 2: Rate limiting + Observability + Networking + Interviewer signals

    Read: interview/rate-limiting/notes.md
    Read: interview/observability/notes.md — critical for senior level
    Read: interview/networking/notes.md — CDN, load balancing
    Read: interview/interviewer-signals/notes.md — read once, no flashcards needed
    Flashcard sessions: 10 rate-limiting + 10 observability + 10 networking
    Written questions: 3 from rate-limiting, 2 from observability
    Design session: "Design a rate limiter for an API gateway"
      Output 1: estimation first — how many API calls/sec?
      Output 2: which algorithm and why, two alternatives ruled out
      Output 3: distributed rate limiting diagram with CDN + LB
      Output 4: what happens when Redis (rate limit store) goes down?
      Output 5: failure — Redis failure means all requests pass or all blocked?
      Output 6 (senior signal): RED metrics, distributed tracing,
                alert on rejection rate spike, V2 = per-region rate limiting
    Deliberate drills: 2 × Constraint Removal (remove the cache, re-solve)

    Sunday test: explain sliding window counter algorithm and why it beats
    fixed window. Then explain what you would monitor on a rate limiter service.

  Week 3: Spanner + classic patterns start (reduced quota)

    Read: Spanner §1 + §2 + §4 (TrueTime)
    Read: interview/classic-patterns/notes.md — URL shortener + news feed sections
    Flashcard sessions: 10 Spanner cards + 10 classic-patterns cards
    Written questions: 2 Spanner + 2 classic patterns
    Design session: "Design Twitter news feed"
      Output 2 must address: fan-out on write vs read, celebrity problem
      After own attempt — use HelloInterview to check.
    Deliberate drills: 2 × Adversarial (attack your own news feed design)

    Sunday test: explain fan-out on write vs fan-out on read, and when
    to use hybrid. No notes.

  Week 4: Spanner consolidation + sharding (reduced quota)

    Read: interview/sharding/notes.md
    Flashcard sessions: 10 Spanner + 10 sharding + 10 classic patterns
    Written questions: Spanner retention questions only, no notes
    Task: Review weak-points.md. Topic 2+ times = 30 min dedicated session.
    Design session: HelloInterview timed session — any storage or feed problem
    Deliberate drills: 2 × Speed Comparison (use new prompts below)

    New speed drill prompts:
      - Token bucket vs sliding window for API rate limiting. 90 seconds.
      - Cache-aside vs write-through for a news feed. 90 seconds.
      - Hash sharding vs range sharding for a time-series DB. 90 seconds.

    Sunday test: explain Bigtable vs Spanner tradeoff without notes.

  Month 1 external checkpoint:
    Explain one classic pattern end-to-end to one person (URL shortener or
    news feed). Include estimation, technology choice, failure handling.
    15-20 minutes. No notes. Record yourself if no person available.

  Month 1 checkpoint criteria:
  - Sunday tests passed for at least 3 of 4 weeks
  - Can do back-of-envelope estimation in under 2 minutes for any problem
  - Answered at least 15 interview questions in writing this month
  - weak-points.md has at least 5 entries (means you were honest)
  - Completed Month 1 external checkpoint

---

## Month 2 — Breadth + Classic Patterns

  Weeks 5-6 quota (full energy — new paper, high engagement):
    3 flashcard sessions + 2 written question answers + 1 design session
    + 1 HelloInterview session + 2 deliberate practice drills

  Weeks 7-8 quota (reduced — second new paper, dip likely):
    2 flashcard sessions + 1 written question answer + 1 design session
    + 1 HelloInterview session + 2 deliberate practice drills

  Week 5-6: Kafka + real-time systems + API design + search systems + chat/notification patterns

    Reading: Kafka §1 → §3 (log structure) → §4 (distribution) → §6
    Read: interview/real-time/notes.md — WebSockets, SSE, presence systems
    Read: interview/api-design/notes.md — REST vs gRPC, pagination, idempotency
    Read: interview/search-systems/notes.md — inverted index, Elasticsearch, autocomplete
    Read: interview/classic-patterns/notes.md — chat + notification sections
    Build Kafka flashcards from gaps.
    Flashcard sessions: 10 Kafka + 10 real-time + 10 search-systems

    Design sessions:
    - Week 5: "Design a chat system (WhatsApp)"
               Output 1: estimation — MAU, messages/day, storage/year
               Output 5: offline delivery + push notifications failure
               Output 6 (senior): WebSocket server count needed, tracing
                         per message, V2 = end-to-end encryption
    - Week 6: HelloInterview timed session — event streaming or chat problem
              Complete debrief template after session

    Deliberate drills week 5: 2 × Constraint Removal
      - Remove WebSockets from chat design. What breaks?
      - Remove Kafka from notification system. What do you use instead?
    Deliberate drills week 6: 2 × Scale Explosion
      - Your chat system handles 10M users. Now 1B. What breaks first?

    Sunday test week 5: explain WebSockets vs SSE vs long polling —
    when to use each. No notes.
    Sunday test week 6: explain message ordering in a chat system and
    what you would monitor in production.

  Week 7-8: Cassandra + distributed transactions + replication + security + object storage (reduced quota)

    Reading: Cassandra §1 → §2 (compare to Bigtable) → §3 (compare to Dynamo)
    Read: interview/distributed-transactions/notes.md — Saga, outbox pattern
    Read: interview/replication/notes.md — sync vs async, replication lag
    Read: interview/availability/notes.md — memorize availability table
    Read: interview/security/notes.md — JWT, OAuth, mTLS, API keys
    Read: interview/object-storage/notes.md — S3, multipart, pre-signed URLs
    Build Cassandra flashcards from gaps.
    Flashcard sessions: 10 Cassandra + 10 distributed-transactions + 10 replication

    Design sessions:
    - Week 7: "Design a web crawler"
               Output 1: estimate pages on internet, crawl rate needed
               Output 6 (senior): monitoring crawl freshness, distributed
                         tracing per URL, V2 = prioritized crawl queue
    - Week 8: HelloInterview timed session — any problem
              Complete debrief template after session

    Deliberate drills week 7-8: 2 × Adversarial each week
      - "Why BFS for crawler? DFS would go deeper faster. Defend."
      - "Your order service uses Saga. What if the orchestrator crashes mid-saga?"

    Sunday test week 7: explain Saga choreography vs orchestration —
    when to use each. Include compensating transaction example.
    Sunday test week 8: given a system with 3 dependencies at 99.9% each —
    combined availability? Downtime per year? What changes at 99.99%?

  Month 2 external checkpoint:
    Design a chat system end-to-end to one person. Include estimation,
    technology choices with ruled-out alternatives, failure handling,
    AND the senior signal (capacity + observability + evolution).
    20 minutes. No notes. Use debrief template to score yourself after.

  Month 2 checkpoint criteria:
  - Sunday tests passed for at least 3 of 4 weeks
  - Completed 4 HelloInterview sessions, each with written debrief
  - Debrief scores improving across sessions
  - Can explain Cassandra as Dynamo + Bigtable synthesis without notes
  - Can produce Output 6 (senior signal) on any design without prompting
  - weak-points.md recurring topics are decreasing, not increasing

---

## Month 3 — Interview Simulation + Full Synthesis

  All weeks quota (simulation phase — consistency over volume):
    2 flashcard sessions (maintenance only) + 1-2 HelloInterview sessions
    + 1 synthesis exercise + 2 deliberate practice drills per week
    (drills in Month 3: Adversarial + Speed only — interview pressure focus)

  Week 9-10: Full synthesis + geo-distributed systems

    Read: interview/geo-distributed/notes.md — active-active vs active-passive, data residency
    Flashcard sessions: 10 geo-distributed cards (spaced with existing reviews)

    Task 1 — internalize the complete system comparison table:

                 Dynamo     Bigtable   Cassandra  Kafka      Spanner
    CAP          AP         CP         Tunable    -          CP
    Data model   KV         Wide-col   Wide-col   Log        SQL+wide
    Range scan   No         Global     Partition  -          Yes
    Transactions No         Single-row No         No         Yes
    Best for     Always     Range+CP   Range+AP   Streams    SQL+scale
                 writable   needed     needed     needed     needed

    Task 2 — without notes, write one paragraph per system:
    "I choose this when X. I do not choose it when Y."

    Task 3 — without notes, write the estimation for 3 classic problems:
    URL shortener, news feed, chat system. QPS + storage + cache size.

    Task 4 — review interviewer-signals/notes.md. For each of the 7
    rubric criteria, write one sentence on how you will demonstrate it.

    HelloInterview: 1 session per week, cross-system problems

    Sunday test week 9: pick any classic pattern. Design it end-to-end
    in 20 minutes with estimation, technology choice, failure handling.
    Sunday test week 10: same but pick a different pattern.

  Week 11-12: Full mock interview simulation

    HelloInterview: 1-2 sessions per week, 45 minutes, fully timed.
    Use the 8-step senior framework. No notes during session.
    After every session: complete the mock interview debrief template.
    Track total score across sessions — it must trend upward.

    Self-practice question bank (not timed, write answers):
    - Bigtable: questions 16, 17, 28, 29, 31
    - Dynamo: questions 11, 18, 20, 30
    - Classic patterns: questions from classic-patterns/interview-questions.md
    - Caching: thundering herd, hot key, cache invalidation questions
    - Design problems: YouTube, distributed job scheduler, bank ledger,
                       notification system at 1B users, global search

    Combined-topics design problems (Month 3 only — these require synthesis):
    These problems intentionally cross multiple domains. You must draw on
    3+ topics per problem. This is what senior interviews actually test.

    - "Design Google autocomplete"
      Topics: search-systems (trie/prefix), caching (hot queries), estimation,
              sharding (by prefix), CDN (edge caching of top-k results)

    - "Design a globally available payment service"
      Topics: distributed-transactions (2PC vs Saga), replication (sync/async),
              geo-distributed (active-passive, data residency), availability (SLO),
              security (API keys, mTLS between services)

    - "Design an e-commerce product search"
      Topics: search-systems (Elasticsearch), caching (popular queries),
              sharding (by product category), object-storage (product images),
              rate-limiting (search API), observability (latency p99)

    - "Design a real-time collaborative document editor"
      Topics: real-time (WebSockets, presence), replication (conflict resolution),
              distributed-transactions (operational transforms vs CRDTs),
              sharding (by document), estimation (concurrent editors)

    - "Design a CDN"
      Topics: networking (GeoDNS, Anycast, edge PoPs), caching (TTL, invalidation),
              object-storage (origin pull from S3), geo-distributed (active-active),
              observability (cache hit rate, origin load)

    Rule for combined-topic problems: all 6 outputs must be present.
    Output 6 must reference at least 2 specific topics (e.g. "I would instrument
    the Elasticsearch scatter-gather with distributed tracing IDs").

    Rule for every design problem in self-practice:
    Output 6 (senior signal) must be present. If it is absent, redo it.

    Deliberate drills week 11-12: all Adversarial + Speed
      - Adversarial: attack every design you produce
        "What breaks at 10x scale?" "What would you monitor first?"
        "What is the first thing you change in V2?"
      - Speed: 90 seconds per comparison, include observability question
        "What do you monitor on a Kafka consumer? 90 seconds."
        "CDN vs no CDN for a news feed API. 90 seconds."

  Month 3 external checkpoint:
    Full 45-minute mock session with one person watching.
    They do not need to be technical — explaining to a non-expert
    surfaces gaps that explaining to a peer does not.
    Score yourself on the 7 rubric criteria after.

  Month 3 checkpoint criteria:
  - Completed at least 4 HelloInterview timed sessions
  - Sunday tests passed every week
  - Can do estimation for any problem before starting design
  - Scored yourself on 7 rubric criteria — no criterion is consistently 0
  - weak-points.md recurring topics resolved or significantly reduced
  - Completed Month 3 external checkpoint

---

## Daily Habit

  Time anchor: pick one fixed time slot and protect it.
  Options: first 20 min of workday, last 20 min before bed, lunch break.
  Unscheduled habits do not happen. Schedule it like a meeting.

  Normal day (20 min):
    10 flashcards from current rotation + 1 written question answer

  Minimum viable day (5 min):
    5 flashcards only. This counts. Streak intact.

  Never skip twice in a row. That is the only rule.

---

## Spaced Repetition Schedule

  After first seeing a flashcard:

    Day 1:   Review
    Day 3:   Review
    Day 7:   Review
    Day 14:  Review
    Day 30:  Review
    Monthly: Maintenance

  Mark cards you get wrong. Review those again the next day.
  A card you keep getting wrong = a gap worth a dedicated session.

---

## Weak Point Tracking

  File: plan/weak-points.md

  First line of that file:
  "If I am reading this after a gap, I am already succeeding.
   The anxious person avoids this file. The serious person opens it."

  Entry format:
    [date] | [topic] | [question failed] | [what was missed]

  Review every Sunday. Same topic 3+ times = 30-min dedicated session
  before continuing the weekly plan.

  Also record weekly progress markers every Sunday:
    1. Which topic was hardest this week?
    2. Which concept felt most natural?
    3. One thing I can explain now that I could not last week.

---

## Recovery Strategy

  Missed one day:    do nothing, continue tomorrow.
  Missed 3-6 days:  use the re-entry ritual, then resume current week.
  Missed one week:  resume at current week, do not go back.
  Missed two weeks: drop back one week, repeat it.
  Missed a month:   restart Month 1 from Week 3 (Spanner) — foundation intact.

  In all cases: do not try to catch up. The plan adapts to you.
  Catching up creates pressure that causes more missed sessions.

---

## Success Criteria (Senior Level)

  Ready for senior interviews when ALL of these are true:

  Knowledge:
  - Sunday test passes for any concept across all papers and topics
  - Can argue "why not X?" for at least 2 alternatives in any design
  - Can explain every design decision in the Raft implementation without code

  Estimation:
  - Can produce QPS + storage + cache size estimate for any problem in 2 min
  - States estimation before drawing any diagram — always, without prompting

  Senior signal:
  - States capacity, observability, and evolution in every design unprompted
  - Debrief template score ≥ 18/24 consistently across last 3 sessions
  - Raises failure modes before interviewer asks about them

  Conversation:
  - Drives the design session — does not wait for interviewer to prompt
  - Recovers cleanly from mistakes mid-interview ("I realize X is wrong, let me fix it")
  - Can handle scope changes without losing structure

  Not ready when:
  - Output 6 (senior signal) requires the interviewer to ask for it
  - Estimation happens after the diagram, not before
  - Technology choice has no ruled-out alternatives
  - Debrief score is flat or declining across sessions
  - Freezes when asked "what would you monitor?" or "what breaks at 10x?"

---

## Notes Reference

  Papers (read in full using the paper reading process):
    Dynamo:    https://www.allthingsdistributed.com/files/amazon-dynamo-sosp2007.pdf
    Bigtable:  https://static.googleusercontent.com/media/research.google.com/en//archive/bigtable-osdi06.pdf
    Spanner:   https://static.googleusercontent.com/media/research.google.com/en//archive/spanner-osdi2012.pdf
    Kafka:     https://notes.stephenholiday.com/Kafka.pdf
    Cassandra: https://www.cs.cornell.edu/projects/ladis2009/papers/lakshman-ladis2009.pdf

  Topic notes (read fully, build flashcards from gaps):
    interview/estimation/notes.md              ← Week 1, critical
    interview/caching/notes.md                 ← Week 1, critical
    interview/rate-limiting/notes.md           ← Week 2
    interview/observability/notes.md           ← Week 2, critical for senior
    interview/networking/notes.md              ← Week 2
    interview/interviewer-signals/notes.md     ← Week 2, read once
    interview/sharding/notes.md                ← Week 4
    interview/classic-patterns/notes.md        ← Weeks 3-8, one pattern/week
    interview/api-design/notes.md              ← Week 5
    interview/real-time/notes.md               ← Week 5
    interview/distributed-transactions/notes.md ← Week 7
    interview/replication/notes.md             ← Week 7
    interview/availability/notes.md            ← Week 7
    interview/consistency/consistency.md       ← Already done
    interview/data-storage/comparision.md      ← Already done
    interview/search-systems/notes.md          ← Month 2 (Week 6), inverted index + Elasticsearch
    interview/security/notes.md                ← Month 2 (Week 7), JWT/OAuth/mTLS
    interview/object-storage/notes.md          ← Month 2 (Week 7), S3, multipart, pre-signed URLs
    interview/geo-distributed/notes.md         ← Month 3, active-active vs active-passive
