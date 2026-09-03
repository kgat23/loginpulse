# AI Usage

## Tools Used

- ChatGPT
- Claude

## How I Used AI

I used AI during this assignment mainly as a development and review tool.

Most of my backend experience is in Java/Spring Boot. I have worked with Go before for concurrent backend systems, but I am less experienced with Go than Java, so I used AI mainly to validate Go-specific implementation choices and review the solution for issues I might have missed.

I used AI for:

- reviewing the PostgreSQL schema and indexing strategy
- discussing raw login events vs pre-aggregated DAU/MAU data
- checking timezone and date boundary handling
- reviewing Go patterns around `context.Context`, `database/sql`, and HTTP handlers
- identifying useful tests and edge cases
- reviewing production and cloud deployment concerns
- reviewing the completed implementation for correctness and unnecessary complexity

The final architecture, schema, API behaviour, and trade-offs were decisions I reviewed and understood before including them in the submission.

---

# Prompt Transcript

## ChatGPT

### 1. Database Design

> I need to build a Go service that records user login events and returns daily and monthly unique user counts. Postgres will be the persistent database.
>
> My initial thought is to keep the raw login events:
>
> `user_logins(id, user_id, login_time)`
>
> and calculate DAU/MAU using `COUNT(DISTINCT user_id)`.
>
> Would you keep raw events as the source of truth or maintain separate daily/monthly aggregate tables? This should be production-aware but is still a 3-5 hour take-home, so I don't want to overengineer it.

### 2. Schema and Indexing

> I think raw events + `COUNT(DISTINCT user_id)` is the better starting point because it keeps one source of truth and avoids consistency issues with aggregates.
>
> The example schema uses `UNIQUE(user_id, login_time)`. I don't think that constraint actually solves duplicate counting because a user can legitimately log in multiple times at different timestamps and `COUNT DISTINCT` already handles that.
>
> Would you keep that constraint?
>
> Also, what index would you use for the daily/monthly queries and where does this approach start becoming expensive?

### 3. Go Project Structure

> Help me decide the project structure before I add more code.
>
> Coming from Spring Boot my instinct is:
>
> ```
> cmd/server
> internal/handler
> internal/service
> internal/repository
> internal/model
> ```
>
> But this is a small Go service and I don't want to recreate Spring architecture with unnecessary layers.
>
> What would an idiomatic but still testable Go structure look like here?

### 4. Database Layer

> I want to implement the Postgres layer using `database/sql`.
>
> The operations I need are roughly:
>
> ```
> RecordLogin(ctx, userID, timestamp)
> GetDailyUniqueUsers(ctx, date)
> GetMonthlyUniqueUsers(ctx, year, month)
> ```
>
> Use parameterized queries and `context.Context`.
>
> Keep it simple. Also explain any Go-specific patterns you use that might not be obvious coming from Java.

### 5. Timezone Handling

> I think timezone handling is probably the easiest place to introduce a subtle correctness bug here.
>
> If timestamps are stored in UTC and the API receives:
>
> `GET /analytics/daily?date=2026-09-01`
>
> I need a clear definition of what that day means.
>
> For this assignment I would rather define all analytics boundaries in UTC instead of introducing configurable timezones.
>
> Would you store login timestamps as `TIMESTAMPTZ`?
>
> Also show me how you would structure the daily/monthly SQL using `[start, end)` timestamp ranges. I would rather avoid wrapping `login_time` in `DATE()` if that makes the timestamp index less useful.

### 6. HTTP API

> I'm thinking of keeping the API minimal:
>
> ```
> POST /logins
> GET /analytics/daily?date=YYYY-MM-DD
> GET /analytics/monthly?month=YYYY-MM
> ```
>
> I want proper input validation and sensible status codes without adding unnecessary abstractions.
>
> Help me implement the handlers while keeping SQL/database logic outside the HTTP layer.

### 7. Tests and Edge Cases

> Before I finish the tests, help me find cases that could make the DAU/MAU results incorrect.
>
> I already have:
>
> - same user logging in multiple times on the same day
> - same user across different days
> - multiple users
> - day boundary
> - month boundary
> - empty result
> - malformed date/month input
>
> What am I missing?
>
> Also separate tests that give useful confidence in my application from tests that would effectively just test PostgreSQL's `COUNT DISTINCT` implementation.

### 8. Scaling

> Assume the table eventually grows to 100M+ login events.
>
> I don't want to implement premature optimizations for this assignment, but I want the README to explain a realistic scaling path.
>
> My current thinking is:
>
> ```
> now:
> raw events + index + COUNT DISTINCT
>
> later:
> partition the event table and/or maintain daily unique-user rollups
>
> much later:
> streaming/analytics infrastructure if ingestion and query volume justify it
> ```
>
> Critique this approach. What am I oversimplifying?

### 9. Production Review

> Review the service assuming multiple instances are running behind a load balancer.
>
> The Go service is stateless and Postgres is the source of truth.
>
> Look specifically for:
>
> - anything that breaks with multiple instances
> - database connection pool issues
> - context/cancellation mistakes
> - graceful shutdown problems
> - query/index problems
> - assumptions that won't hold at higher traffic
>
> Don't rewrite the project. Just point out the problems and explain why they matter.

### 10. Final Review

> Review the completed repository as if you were the engineer evaluating this take-home.
>
> Don't rewrite the project.
>
> Look for:
>
> 1. correctness bugs
> 2. non-idiomatic Go
> 3. SQL/indexing problems
> 4. missing error handling
> 5. missing tests
> 6. unnecessary complexity
> 7. anything claimed in the README that the implementation doesn't actually do
>
> Rank the issues by severity.

---

## Claude

### 1. Independent Code Review

> Review this completed Go take-home as a backend engineer.
>
> The service records user login events in PostgreSQL and provides daily and monthly unique-user counts.
>
> Don't redesign the project just because you would structure it differently.
>
> Look specifically for actual correctness, SQL, concurrency, API, testing, or production-readiness problems. Pay particular attention to duplicate logins and UTC day/month boundaries.
>
> Separate actual issues from optional improvements.

### 2. Simplification Review

> Now review the same implementation specifically for overengineering.
>
> This is supposed to be a 3-5 hour backend assignment, not a full analytics platform.
>
> Point out abstractions, packages, or infrastructure that aren't buying meaningful correctness, readability, testability, or deployment value.

### 3. Interview Review

> Assume I have to walk through this repository in an interview without using AI.
>
> Give me the 10 hardest questions you would ask about the implementation, SQL schema, indexing, timezone decisions, testing, and scaling trade-offs.
>
> Don't give me the answers yet.

---

## Final Note

AI was most useful for checking Go-specific implementation details and acting as a second reviewer.

I intentionally kept the submitted solution relatively small rather than adding infrastructure that the current requirements do not need.

I reviewed AI-generated suggestions before applying them and can explain the code and the decisions in the submission, including the schema, uniqueness semantics, SQL queries, indexing, UTC handling, API behaviour, tests, and how I would evolve the design at higher scale.