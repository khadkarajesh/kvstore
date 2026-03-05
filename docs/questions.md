1. What does the leader track per follower? (nextIndex, matchIndex)
2. What exactly does prevLogIndex/prevLogTerm prove?
These are used to check the consistency. In Raft algorithm, if the previous index and term has the same value means all the previous values before the prevLogIndex and prevLogTerm have the safe value.
3. What does the follower do when the check fails?
When check fails, it responds with failure and Leader retries the appendEntry until it finds the matching index and term which means follower will reponds with success.
4. What does the follower do when the check passes?
When check passes it will add the rest of the value from Leader to the state machine.
5. When is an entry considered "committed"?
Entry is considered committed when the leader writes the value in its state machine, which will done when all followers writes the latest value into their state machine.