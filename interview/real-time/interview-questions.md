## Real-Time Data Delivery — Interview Questions

These are open-ended questions as they would appear in a real system design interview.
Answer each out loud or in writing before checking your notes.

---

### Protocol Selection

1. What is the first clarifying question you ask when an interviewer says "design a real-time notification system"? How does the answer change your protocol choice?

2. Compare short polling, long polling, SSE, and WebSockets. For each, state the mechanism in one sentence and give one scenario where it is the right choice.

3. A live sports score dashboard shows score updates to users. The client never sends data — it only receives. Which protocol do you choose, and why is WebSocket more than you need here?

4. Google Docs requires that when User A types a character, User B sees it within 100ms. Which protocol do you use? What happens if you use SSE instead?

5. An engineer says "we should use WebSockets for everything real-time." What are two scenarios where WebSockets are the wrong choice, and what would you use instead?

6. A corporate proxy in the user's network blocks all non-standard HTTP upgrades. Your WebSocket connections are being dropped. What protocol do you fall back to, and what is the latency trade-off?

---

### Scaling

7. User A is connected to WebSocket Server 1. User B is connected to WebSocket Server 2. User A sends a message to User B. What is the problem, and what is the canonical solution?

8. Walk through the pub/sub fan-out architecture for WebSocket servers. What role does Redis play? What is the latency cost of the pub/sub hop?

9. How many concurrent WebSocket connections can a single server handle? If you have 100M DAU with 10% online at peak, how many WebSocket servers do you need?

10. Your WebSocket servers are stateful — each holds open connections. A user's server is taken out of service for a deploy. What happens to that user's connection, and how does the client handle reconnection?

11. What are two approaches to routing WebSocket connections — consistent hashing vs. any-server with pub/sub? What are the trade-offs of each for a chat application?

---

### Presence Systems

12. Design a presence system (is this user online?) using Redis. Walk through the write path when a user connects, the heartbeat mechanism, and the read path when another user queries their status.

13. A user's mobile app is backgrounded by the OS. The WebSocket connection is still technically open, but the app is frozen. How does your presence system eventually detect this user as offline?

14. A user has two browser tabs open on the same account. They close one tab. How does your presence system correctly keep them shown as "online" rather than immediately marking them offline?

15. A Slack-like workspace has 500 users in a channel. When any user changes presence status, all other 499 users must be notified. Walk through the fan-out: how many messages does this generate, and how does Redis pub/sub help?

---

### Fan-out

16. Twitter has 100M followers for a celebrity. The celebrity posts a tweet. How do you fan out this tweet to all 100M followers in real-time? Walk through the write fan-out vs. read fan-out trade-off.

17. A chat room has 10,000 participants. One user sends a message. How does the server deliver it to all 10,000 WebSocket connections efficiently? What is the bottleneck?

18. You are designing push notifications for a mobile app with 50M users. When a news event occurs, you must notify all 50M users within 2 minutes. Walk through the architecture, including the fan-out layer.

19. What is the difference between push fan-out (write-time) and pull fan-out (read-time)? For a high-follower-count social media post, which do you use and what is the threshold for switching strategies?

20. A message in a group chat must be delivered to 200 participants, each on a different WebSocket server across a fleet of 50 servers. Walk through how pub/sub routing ensures delivery without requiring every server to know about every user.

---

### Design Problems

21. Design WhatsApp message delivery. Cover: protocol choice, how a message gets from sender to recipient, what happens when the recipient is offline, and how you scale to 100M concurrent connections.

22. Design the real-time "typing indicator" feature for a chat app. A user sees "{name} is typing..." when someone in their chat is typing. How does this work at the protocol level, and how do you avoid flooding the server with keystroke events?

23. Design a live leaderboard for a mobile game. Scores update every few seconds for 1M concurrent players. The leaderboard must show the top 100 in real time. Walk through data flow from score write to leaderboard update to client display.

24. An interviewer says "design a stock price ticker that shows real-time prices to 5M concurrent browser users." Which protocol do you use, how do you scale the fan-out, and what do you do differently for the 1,000 users watching high-frequency trading data vs. the 4.99M watching once a minute?

25. You designed a chat system with WebSockets and now the interviewer says "how does this work on mobile?" Walk through battery, background, and unreliable network constraints, and how they change your real-time delivery strategy for mobile vs. browser clients.
