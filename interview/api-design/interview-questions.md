## API Design — Interview Questions

These are open-ended questions as they would appear in a real system design interview.
Answer each out loud or in writing before checking your notes.

---

### REST vs gRPC vs GraphQL

1. You are designing a new service. The interviewer asks you to define the API. What is your first question, and how does the answer change which protocol you choose?

2. What are the four communication patterns gRPC supports? Give a concrete use case where each is the right choice.

3. A mobile client needs to display a user's profile: name, avatar, follower count, and their last 3 posts. With REST you need two endpoints and two round trips. How does GraphQL solve this, and what new problem does it introduce?

4. Your team has five services: auth, user, order, inventory, and notification. Which of those would you expose with REST and which with gRPC? Justify each.

5. An engineer argues that GraphQL solves REST's over-fetching problem. What is the N+1 query problem, how does GraphQL create it, and what is the standard fix?

6. gRPC uses Protocol Buffers. What are two concrete advantages of Protobuf over JSON, and what is one operational cost you pay for using them?

---

### Pagination & Versioning

7. You are paginating a feed of 100 million posts ordered by created_at. An engineer proposes offset pagination with `?page=5&per_page=20`. What are the two specific failure modes at scale, and what do you propose instead?

8. Walk through how cursor-based pagination works at the database query level. What does the cursor encode, and why is it stable under concurrent inserts?

9. Twitter, Stripe, and GitHub all use cursor-based pagination. What is the O(n) cost problem with offset pagination that forces this choice at scale?

10. You need to introduce a breaking change to your REST API — removing a field that existing clients depend on. What are the two main versioning strategies, when do you choose each, and what do you tell clients?

11. gRPC does not use URL versioning. How does it handle backward compatibility over time, and what rule must you follow when changing a `.proto` file?

---

### Idempotency & Safety

12. A client sends a POST /payments request, gets a network timeout, and retries. What happens without idempotency guarantees? Walk through the full idempotency key solution including what is stored and where.

13. Which HTTP verbs are safe? Which are idempotent? Give the correct answer for GET, POST, PUT, PATCH, and DELETE.

14. Your team ships a payment API. An engineer says "we'll just check for duplicate requests by looking at the amount and timestamp." What is wrong with this approach?

15. The idempotency key is stored in Redis with a 24-hour TTL. A client retries after 25 hours. What should happen, and is this correct behavior?

16. You are building an email notification service. A saga step sends a welcome email. The orchestrator crashes and replays the step. How do you prevent sending the email twice?

---

### Rate Limiting & Auth

17. You are building a public REST API. What four dimensions can you apply rate limiting on, and what are the trade-offs of each?

18. An API returns HTTP 429. What headers should you include in the response, and what should a well-behaved client do when it receives 429?

19. Where in the architecture does rate limiting live — API gateway vs. service layer vs. client side? When is each appropriate?

20. An engineer proposes using API keys for authentication on a public-facing endpoint. What do API keys give you that stateless JWT tokens do not, and vice versa?

---

### Real-world Design

21. Design the API for a URL shortener. What endpoints do you define, what HTTP verbs do you use, and how do you handle the redirect? Be specific about status codes.

22. You are designing a chat application API. Should the message-sending endpoint be REST or WebSocket? What about reading message history?

23. An interviewer asks you to design the Twitter timeline API. Walk through: which protocol, what pagination strategy, how you handle versioning, and what idempotency concern exists for the tweet POST endpoint.

24. You are designing an internal service where ServiceA calls ServiceB 50,000 times per second. What API protocol do you choose and why? What happens if you chose REST instead?

25. A senior engineer reviews your design and says: "You forgot to handle client retries safely." What does this mean, what can go wrong, and how do you fix it before the design review ends?
