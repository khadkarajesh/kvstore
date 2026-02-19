# Raft Consensus — Knowledge Gap Reference

> Written for: System Design Interview Prep + Distributed KV Store Project
> Revisit this when you feel confused about elections, failures, or test design.

---

## Table of Contents

1. [How to Think About Tests](#1-how-to-think-about-tests)
2. [RequestVote Decision Table](#2-requestvote-decision-table)
3. [The Log Up-To-Date Check](#3-the-log-up-to-date-check)
4. [Why Terms Never Roll Back](#4-why-terms-never-roll-back)
5. [Failure Taxonomy](#5-failure-taxonomy)
6. [Minority Partition — Full Walkthrough](#6-minority-partition--full-walkthrough)
7. [Idempotent Re-Vote](#7-idempotent-re-vote)
8. [Test Cases Map to Real Scenarios](#8-test-cases-map-to-real-scenarios)

---

## 1. How to Think About Tests

### The Gap
You were writing tests by reading your own code — tracing what paths exist
and trying to cover them. This is the hard direction.

### The Fix
Work from the **spec**, not the code. Every condition in the Raft paper is
a test case. You don't invent them — you derive them.

```
WRONG approach:                     RIGHT approach:
  Read code                           Read Raft paper rule
  → find branches                     → ask "what breaks if this rule
  → write test to hit each branch       didn't exist?"
                                      → write test that would fail
                                        without that rule
```

### The Rule
Before writing any test, write the rule in plain English first:

```
Rule (§5.4.1): A server must refuse to vote if the candidate's log
               is less up-to-date than the server's own log.

Plain English: Even if the term is newer, a stale-log candidate
               cannot win an election.

Test: Set up server with entries, send candidate with empty log,
      expect VoteGranted=false.
```

The test almost writes itself from there.

---

## 2. RequestVote Decision Table

A server grants a vote **only if ALL three conditions hold**:

```
┌─────────────────────────────────┬─────────────┬───────────────┐
│ Condition                       │ Value       │ Result        │
├─────────────────────────────────┼─────────────┼───────────────┤
│ 1. req.Term >= currentTerm      │ true        │ continue      │
│                                 │ false       │ DENY (§5.1)   │
├─────────────────────────────────┼─────────────┼───────────────┤
│ 2. votedFor == 0                │ true        │ continue      │
│    OR votedFor == candidateId   │ true        │ continue      │
│    (already voted for other)    │ false       │ DENY (§5.2)   │
├─────────────────────────────────┼─────────────┼───────────────┤
│ 3. Candidate log up-to-date     │ true        │ GRANT         │
│                                 │ false       │ DENY (§5.4.1) │
└─────────────────────────────────┴─────────────┴───────────────┘
```

**Important**: Conditions 2 and 3 are independent. Both must pass.
Seeing a higher term (condition 1) does NOT automatically grant the vote —
it only resets votedFor and steps the node down. The log check still runs.

### Complete Test Matrix

| Test | Term | Voted? | Log Fresh? | Expected |
|------|------|--------|------------|----------|
| RejectStaleTerm | old | — | — | DENY |
| GrantVoteFirstTime | same | no | yes (empty=empty) | GRANT |
| DenyDoubleVote | same | yes (other) | yes | DENY |
| UpdateTermOnNewerTerm | newer | resets to 0 | no (stale) | DENY vote, UPDATE term |
| StaleLogDenied_OlderTerm | same | no | no (term behind) | DENY |
| StaleLogDenied_ShorterIndex | same | no | no (index behind) | DENY |
| IdempotentReVote | same | yes (same candidate) | yes | GRANT |

---

## 3. The Log Up-To-Date Check

### The Rule (Raft §5.4.1)
> "If the logs have last entries with different terms, then the log with
> the later term is more up-to-date. If the logs end with the same term,
> then whichever log is longer is more up-to-date."

### Two Sub-Cases

```
Sub-case 1: Different lastLogTerm
──────────────────────────────────
  Server:    lastLogTerm=3, lastLogIndex=5
  Candidate: lastLogTerm=2, lastLogIndex=9   ← longer but older term

  Result: DENY — term wins, index doesn't matter

Sub-case 2: Same lastLogTerm, different lastLogIndex
────────────────────────────────────────────────────
  Server:    lastLogTerm=2, lastLogIndex=5
  Candidate: lastLogTerm=2, lastLogIndex=3   ← same term, shorter log

  Result: DENY — tie-break by index
```

### In Code
```go
logUpToDate := req.LastLogTerm > s.lastLogTerm ||
    (req.LastLogTerm == s.lastLogTerm && req.LastLogIndex >= s.lastLogIndex)
```

### Why This Matters
This is the only rule that prevents a node with a stale log from winning
an election and overwriting committed entries.

```
5-node cluster, leader committed 10 entries to nodes 1,2,3
Leader crashes.
Node 4 has 3 entries, Node 5 has 3 entries.

Node 4 campaigns:
  → Node 1 rejects (log check fails: 3 < 10)
  → Node 2 rejects (log check fails: 3 < 10)
  → Node 3 rejects (log check fails: 3 < 10)
  → Node 5 grants  (both have 3 entries)
  Node 4 can't win quorum → election fails ✓

Node 1 campaigns:
  → Node 2 grants (log matches)
  → Node 3 grants (log matches)
  Node 1 wins with 3 votes → committed data is safe ✓
```

---

## 4. Why Terms Never Roll Back

### The Question
Node 3 was partitioned, kept incrementing its term (1→2→3→...→9).
Its election at term=9 fails (no quorum). Why does it keep term=9
instead of rolling back to term=1?

### The Answer — Terms Are a Logical Clock

```
Term does NOT mean: "I won an election for this term"
Term MEANS:         "I am aware the cluster has reached this era"
```

### What Already Happened Can't Be Undone

When Node 3 sent RequestVote(term=9), other nodes already updated:

```
Node 3 sends RequestVote(term=9)
        │
        ├──▶ Node 1: currentTerm 1 → 9  ← side effect already happened
        └──▶ Node 2: currentTerm 1 → 9  ← side effect already happened

Node 3 loses (stale log)
Node 3 rolls back to term=1  ← hypothetical

Now:
  Node 1: term=9
  Node 2: term=9
  Node 3: term=1  ← permanently isolated
                     every message from Node 3 gets rejected as stale
```

### Rolling Back Breaks Safety

If terms could be reused:
```
Round 1: Node 3 campaigns for term=10, loses
Node 3 rolls back to term=9
Round 2: Node 3 campaigns for term=10 again

Problem: Node 1 already reset votedFor after seeing term=10 in round 1.
         Node 1 could vote for Node 3 in "round 2's term=10"
         AND it already voted for someone else in "round 1's term=10"
         → Two different leaders in term=10 → Split Brain
```

### The Mental Model
```
Terms are like calendar years.

2024 happened. You can't un-happen 2024
just because nothing important occurred in it.
2025 must come after 2024, regardless.

A failed election is just a year where no leader
was elected. The clock still moved forward.
```

---

## 5. Failure Taxonomy

### Category 1: Node Crash

| Sub-scenario | What breaks | Raft's response |
|-------------|-------------|-----------------|
| Follower crashes | Leader stops replicating to it | Resume when it comes back, catch up via AppendEntries |
| Leader crashes | No heartbeats sent | Followers detect timeout → new election |
| Candidate crashes mid-election | Votes in-flight, no quorum | Election times out → someone else retries |
| Node crashes after voting | `votedFor` lost (in-memory) | **Can double-vote on restart** → requires disk persistence |

> The last row is why Phase 3 (Persistence) matters. `votedFor` and
> `currentTerm` MUST be persisted to disk before responding to any RPC.

### Category 2: Network Partition

| Sub-scenario | What breaks | What happens |
|-------------|-------------|--------------|
| Minority partition | Isolated node keeps timing out | Term inflates; disrupts cluster on rejoin |
| Split brain (even split) | 2 leaders possible | Neither gets quorum (odd cluster sizes prevent this) |
| Leader partitioned | Old leader still accepts writes | New leader elected; old leader's writes lost |
| Message delay | Stale votes arrive late | Ignored — term check rejects them |

> Always use **odd cluster sizes** (3, 5, 7). Even splits create the
> possibility of two halves each thinking they have quorum.

---

## 6. Minority Partition — Full Walkthrough

This scenario directly maps to two of your test cases.

### Phase 1: Healthy Cluster

```
┌─────────┐     heartbeat     ┌─────────┐
│ Node 1  │──────────────────▶│ Node 2  │
│ LEADER  │                   │FOLLOWER │
│ term=1  │──────────────────▶│ term=1  │
└─────────┘     heartbeat     └─────────┘
     │
     │ heartbeat
     ▼
┌─────────┐
│ Node 3  │
│FOLLOWER │
│ term=1  │
└─────────┘
```

### Phase 2: Node 3 Partitioned

```
┌─────────┐                   ┌─────────┐
│ Node 1  │──────────────────▶│ Node 2  │
│ LEADER  │   still healthy   │FOLLOWER │
│ term=1  │                   │ term=1  │
└─────────┘                   └─────────┘

[partition wall]

┌─────────┐
│ Node 3  │  ←── no heartbeats received
│FOLLOWER │
│ term=1  │
└─────────┘
```

### Phase 3: Node 3 Campaigns Alone

```
Node 3 timer fires → term=2 → RequestVote(term=2)
                                      │
                    ╳ blocked         ├──▶ (silence)
                                      └──▶ (silence)

Node 3 can't reach quorum (needs 2, only has itself)
Timer fires again → term=3 → same result
...
Timer fires again → term=9 → same result
```

Node 1 and Node 2 are unaware. Cluster operates normally.

### Phase 4: Partition Heals

```
Node 1 sends heartbeat(term=1) to Node 3
Node 3 (term=9) receives it:
  req.Term(1) < currentTerm(9) → REJECT
  Node 3 returns term=9 in response

Node 1 receives response with term=9:
  9 > 1 → Node 1 steps down immediately
  Node 1: term=9, Follower
  Node 1 stops sending heartbeats
```

### Phase 5: Forced Re-election

```
Node 1: term=9, Follower (just stepped down)
Node 2: term=1, Follower (hasn't seen term=9 yet)
Node 3: term=9, Candidate (just rejoined, log=empty)

Nobody sending heartbeats → all timers start ticking
```

**Path A — Node 2 fires first:**
```
Node 2 campaigns for term=2
  → Node 1 rejects (1 < 9), returns term=9
  → Node 3 rejects (1 < 9), returns term=9
Node 2 sees term=9, updates, becomes Follower

All three at term=9. Node 1 fires:
Node 1 campaigns for term=10 (has full log)
  → Node 2 grants ✓ (log is up to date)
  → Node 3 grants ✓ (log is up to date)
Node 1 wins
```

**Path B — Node 3 fires first:**
```
Node 3 campaigns for term=10 (log=empty)
  → Node 1 rejects (log check: 0 < full log)  ← YOUR MISSING TEST
  → Node 2 rejects (log check: 0 < full log)
Node 3 fails. All at term=10.

Node 1 fires → campaigns for term=11
  → wins because log is up to date
```

### Key Insight

```
Term inflation (Phase 3) → tested by: UpdateTermOnNewerTerm
Stale log rejection (Phase 5, Path B) → tested by: StaleLogDenied_*

These two tests cover the SAME real-world event.
They're different sides of the minority partition scenario.
```

---

## 7. Idempotent Re-Vote

### The Scenario

```
Normal election, term=1:

Node 2 campaigns.
Node 1 votes for Node 2 → votedFor=2, sends VoteGranted=true
Response packet is DROPPED by the network.

Node 2 never received the response → retries RequestVote to Node 1.
Node 1 receives the same RequestVote again.
```

### What Must Happen

```
Node 1 state: term=1, votedFor=2

Receives: RequestVote(term=1, candidateId=2)

alreadyVoted = votedFor(2) != 0 && votedFor(2) != candidateId(2)
             = true && false
             = false   ← NOT considered "already voted for someone else"

→ Vote GRANTED again (idempotent)
```

### Why This Is a Safety Rule, Not a Convenience

If Node 1 denied the retry:
```
Node 2 sent RequestVote to all 3 nodes.
Node 1: voted, response dropped, now denies retry
Node 3: voted, response received

Node 2 only confirmed 1 vote (Node 3).
Needs 2 to win. Election fails.
New election starts. Another node wins.

But Node 2 might have actually won the first election
(Node 1 DID vote for it) — the cluster just doesn't know.
```

Idempotent re-vote ensures **network unreliability doesn't stall elections**.

### What votedFor=0 Means vs votedFor=candidateId

```
votedFor=0          → haven't voted yet this term, free to vote
votedFor=2          → voted for candidate 2, ONLY candidate 2 can re-request
votedFor=1 (other)  → voted for someone else, all others denied
```

---

## 8. Test Cases Map to Real Scenarios

| Test | Real cluster scenario |
|------|-----------------------|
| `RejectStaleTerm` | Partitioned node rejoins with old term, campaigns |
| `GrantVoteFirstTime` | Normal election, fresh follower votes |
| `DenyDoubleVote` | Two candidates in same term, follower already committed vote |
| `UpdateTermOnNewerTerm` | Leader/Follower sees higher term → must step down (minority partition rejoin) |
| `StaleLogDenied_OlderTerm` | Node missed elections, has older lastLogTerm, tries to become leader |
| `StaleLogDenied_ShorterIndex` | Node partially replicated, same term but shorter log |
| `IdempotentReVote` | Network drops vote response, candidate retries |

---

## Quick Reference Cheatsheet

```
RequestVote — when to GRANT:
  req.Term >= currentTerm
  AND (votedFor == 0 OR votedFor == candidateId)
  AND (req.LastLogTerm > lastLogTerm
       OR (req.LastLogTerm == lastLogTerm AND req.LastLogIndex >= lastLogIndex))

On seeing req.Term > currentTerm (any RPC, not just RequestVote):
  currentTerm = req.Term
  votedFor    = 0
  state       = Follower

Terms:
  Only move forward, never back.
  Failed election → keep the incremented term.
  Term = logical clock, not vote counter.

Failures:
  Node crash  → election timeout detects it
  Partition   → same mechanism, but isolated node inflates term
  Both        → log check prevents stale node from winning
```

---

*Reference: Raft paper — "In Search of an Understandable Consensus Algorithm" by Ongaro & Ousterhout*
*Sections: §5.1 (Terms), §5.2 (Leader Election), §5.4.1 (Election Restriction)*
