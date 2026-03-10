Interviewer Signals — How FAANG System Design Interviews Are Scored

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
What Interviewers Score (The Rubric)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Criterion 1 — Problem Clarification
    Did you ask the right questions before starting to design?
    Good: "What is the read/write ratio? Do we need real-time or is eventual consistency acceptable? What is the expected QPS and storage requirement?"
    Bad: jumping to architecture before establishing requirements.
    Signal: interviewers are checking whether you default to assumptions or gather data first.

  Criterion 2 — Estimation
    Did you quantify the scale before choosing technology?
    Good: "100M users, 50 writes/user/day → 60K writes/sec → need multiple DB shards"
    Bad: "We'll use Cassandra because it's scalable."
    Signal: "scalable" is meaningless without a number. Technology choice must follow from a calculation.

  Criterion 3 — Structured Progression
    Did you follow a logical order: requirements → estimation → high-level design → deep dives?
    Good: explicitly naming each phase as you enter it.
    Bad: jumping between components randomly, leaving gaps.
    Signal: interviewers write down whether your design covered all major components.

  Criterion 4 — Technology Justification
    Did you explain why you chose each technology?
    Good: "I'm using Kafka here instead of a direct DB write because the notification fan-out is O(followers) — we can't block the write path on that."
    Bad: "I'll use Kafka because Kafka is good for streaming."
    Signal: "why" answers demonstrate understanding. "what" answers demonstrate name-dropping.

  Criterion 5 — Tradeoff Awareness
    Did you raise downsides before being asked?
    Good: "Fan-out on write is fast to read but expensive for celebrities. I'll add the hybrid approach."
    Bad: only addressing tradeoffs when the interviewer asks "what's the downside of that?"
    Signal: self-raising tradeoffs is a senior-level behavior. Waiting to be asked is junior.

  Criterion 6 — Failure Handling
    Did you address what happens when components fail?
    Good: "If the notification worker crashes after sending but before marking delivered, it will retry. I'll use an idempotency key to prevent duplicate sends."
    Bad: designing a system that works only on the happy path.
    Signal: every component you draw → ask yourself "what if this fails?" and answer it.

  Criterion 7 — Driving the Conversation
    Did you lead the interview, or did the interviewer have to pull answers out of you?
    Good: after each component, propose next steps, ask for feedback, signal where to go deeper.
    Bad: answering only what is asked, waiting for the interviewer to direct every transition.
    Signal: interviewers note whether the candidate drove or was driven. Leadership in design = leadership on the job.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
How to Handle "I Don't Know"
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Never say: "I don't know" and stop.
  Never say: "I haven't used that" and wait for the interviewer to move on.

  Do say:
    "I haven't worked with X directly, but reasoning from first principles — [what you know about the problem domain and constraints] — I would expect the solution to [reasoning]."

  Why reasoned wrong answers score higher than silence:
    The interview is measuring problem-solving ability, not a knowledge quiz.
    A candidate who reasons correctly to a wrong answer shows:
      → structured thinking
      → willingness to engage with hard problems
      → ability to operate in uncertainty
    A candidate who goes silent shows none of those things.

  Example — asked about "how Zookeeper leader election works":
    "I haven't implemented Zookeeper's algorithm specifically, but I know leader election
     needs to handle split-brain and network partitions. Reasoning from Raft: you need
     a quorum of nodes to agree on a leader, with fencing to prevent a deposed leader
     from acting. Zookeeper uses ephemeral nodes — each candidate creates an ephemeral
     sequential node; whoever creates the lowest-numbered node wins. Ephemeral: if the
     leader crashes, ZK deletes the node and the next candidate takes over."
    → This is a correct answer derived by first-principles reasoning from related knowledge.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
How to Drive the Conversation
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Opening (after clarifying requirements):
    "I'm going to start with a high-level design and then we can dive into whichever
     component you're most interested in. Does that work?"
    → gives interviewer agency while establishing that you are leading.

  Transitioning between components:
    "I've covered the write path. I've made a fan-out on write decision with a hybrid for celebrities.
     Should I go deeper on the feed assembly and caching, or move on to the storage layer?"
    → names what you've done, names the open tradeoff, asks for direction.

  Raising failure modes unprompted:
    "Before I move on — one failure case worth addressing here: if the message queue goes down,
     the notification worker stops consuming but upstream services keep producing.
     I'd add dead letter queuing and an alert on consumer lag. OK to continue?"
    → never wait for "what about failures?" to be asked.

  Time management:
    45-minute interview, typical breakdown:
      5 min  → clarification + requirements
      5 min  → estimation
      15 min → high-level design (all major components, happy path)
      15 min → deep dive (2-3 components, failure handling, tradeoffs)
      5 min  → questions + wrap-up
    If 25 minutes in and you have not drawn a complete high-level design → you are running behind.
    Signal: tell the interviewer explicitly: "I want to make sure I cover all components at a high level
     before diving deep. Let me finish the overview quickly and then we'll go deeper."

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Senior vs Junior Signals
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Junior: describes what a system does.
  Senior: explains why the system makes that choice and what it gives up.

  Junior: "I'll use Redis for the cache."
  Senior: "I'll use Redis here because this data is session-scoped, small, and read-heavy.
           I'm accepting the risk of a brief stale window between write and cache invalidation —
           that's fine because session data is user-specific and the staleness window is <1s."

  ---

  Junior: waits to be asked about failures.
  Senior: proactively raises failure modes and addresses them before being asked.

  ---

  Junior: picks one technology and defends it.
  Senior: picks a technology, names two alternatives, explains why each was ruled out.

  Example:
    Junior: "I'll use Cassandra."
    Senior: "I need a write-heavy, wide-column store with geographic distribution.
             Cassandra is the primary choice. DynamoDB would work but ties us to AWS.
             HBase would work but requires a Hadoop ecosystem we don't have.
             Cassandra gives the right tradeoffs for this use case."

  ---

  Junior: describes a happy-path architecture.
  Senior: identifies the hardest part of the problem in the first 5 minutes.

  Example (news feed):
    Junior: "User posts → store in DB → followers retrieve."
    Senior: "The hardest part here is fan-out for celebrity accounts. Before I design
             the happy path, let me establish how I'll handle a user with 100M followers
             without blocking the write path for minutes."

  ---

  Junior: designs first, estimates later (or never).
  Senior: estimates first, then lets the numbers drive the design.

  Example:
    Before drawing: "Let's assume 500M users, 200M DAU, 100 writes/user/day.
     That's about 230K writes/sec. At that rate a single SQL primary hits its limit →
     I'll need sharding. Let me design around that constraint."

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
How to Handle Scope Creep
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  What scope creep looks like:
    You've designed a URL shortener. Interviewer says: "Oh, and it needs analytics."
    You've designed a chat system. Interviewer says: "Also, it needs end-to-end encryption."

  Wrong response: silently absorb the new requirement and try to retrofit it.
    → blows up your existing design
    → wastes time on a change you weren't asked to make in depth
    → looks like you can't maintain scope discipline

  Right response:
    "That's an important requirement. Adding analytics changes the redirect flow —
     I'd need to use 302 instead of 301 (browser caching would hide clicks from us),
     and add an async click event to a logging pipeline. Should I redesign the redirect
     and storage components now, or note this as a follow-up and continue with the
     current scope?"
    → names the specific impact
    → gives the interviewer a choice
    → demonstrates you understand the ripple effect

  If interviewer says "note it and continue": write it on the whiteboard explicitly.
  If interviewer says "redesign now": acknowledge scope change, adjust time accordingly.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
The One Thing That Fails Most Candidates
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Not failure handling. Not a wrong technology choice. Not poor knowledge of databases.

  It is: jumping to architecture before quantifying the problem.

  What it looks like:
    Question: "Design a URL shortener."
    Candidate: "OK, so we'll have an API server that takes the long URL, we'll use a UUID
                generator, store it in MySQL, put Redis in front for caching..."
    No clarification. No estimation. No QPS number. No storage number.

  What the interviewer thinks:
    "This person is pattern-matching to a solution they've seen, not solving this problem."
    "How would they know whether MySQL scales if they don't know the QPS?"
    "At scale, intuition breaks down. This candidate doesn't think at scale."

  The fix — always in this order:
    1. Clarify requirements (5 min)
       "What is the expected number of daily active users? Read/write ratio? Analytics?"
    2. Estimate scale (5 min)
       "At 100M URLs/day that's ~1,200 writes/sec, 115K redirects/sec. Storage: ~18TB/year."
    3. Let the numbers dictate the design
       "At 115K reads/sec, a single MySQL instance won't cut it → I need caching and read replicas
        at minimum, possibly sharding. Let me design for that."

  Interviewers literally wait to see if you ask for scale before they volunteer it.
  If you assume small scale and design accordingly → scored as not thinking at scale.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
How to Recover from a Mistake Mid-Interview
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  You are 20 minutes in. You realize your design has a flaw.

  Wrong response: hide it, hope the interviewer doesn't notice, keep going.
    → interviewers notice. They are experts. Hiding flaws looks like you can't evaluate your own work.

  Right response:
    "Actually, I realize there's a problem with what I designed for the message queue.
     I set the consumer to commit the offset before calling the notification provider.
     If the provider fails after commit, we lose the message — no retry possible.
     Let me fix that: commit the offset only after the provider confirms delivery,
     and use idempotency keys to make retries safe."

  Why self-correction scores well:
    → It shows you understand the system deeply enough to find your own bugs.
    → It shows intellectual honesty.
    → It shows you would do this on the job, not ship broken designs.

  Interviewers regularly note in feedback: "candidate caught and corrected their own mistake —
   strong signal of systems thinking."

  Rule: never hide a flaw you've found. Always name it and fix it.

