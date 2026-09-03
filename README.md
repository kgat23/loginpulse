# loginpulse

Tracks user login events and reports daily/monthly unique user counts.

## Running

```bash
docker compose up --build
```

This starts Postgres (schema auto-applied from `migrations/`) and the server on `:8080`.

To run against a local Go toolchain instead:

```bash
export DATABASE_URL="postgres://analytics:analytics@localhost:5432/analytics?sslmode=disable"
go run ./cmd/server
```

Run tests (no database required — they use an in-memory store):

```bash
go test ./...
```

## API

```
POST /logins
  {"user_id": "<uuid>", "login_at": "2026-09-02T10:00:00Z"}   # login_at optional, defaults to now

GET /analytics/daily?date=2026-09-02
  {"date": "2026-09-02", "unique_users": 2}

GET /analytics/monthly?month=2026-09
  {"month": "2026-09", "unique_users": 3}
```

Example:

```bash
curl -X POST localhost:8080/logins \
  -d '{"user_id":"11111111-1111-1111-1111-111111111111","login_at":"2026-09-02T08:00:00Z"}'

curl 'localhost:8080/analytics/daily?date=2026-09-02'
# {"date":"2026-09-02","unique_users":1}
```

Sample data with expected outputs: [testdata/sample_logins.sql](testdata/sample_logins.sql).

## Design decisions

**Timezone**: all bucketing ("what day/month is this login in") is done in UTC. `login_at`
is stored as `TIMESTAMPTZ`, converted to UTC on ingest, and truncated to a UTC calendar
day/month. This makes bucketing deterministic regardless of where the request came from,
and is the assumption to revisit first if "day" should instead mean the user's local day
(it would require storing a per-user or per-request timezone and bucketing per-timezone,
which fragments a single "daily count" into one per timezone).

**Uniqueness / no double-counting**: ingestion writes to three tables in one transaction:
the raw `user_logins` log, and two dedup tables (`daily_active_users`,
`monthly_active_users`) keyed on `(bucket, user_id)` with `ON CONFLICT DO NOTHING`. A user
logging in 50 times in a day produces 50 raw rows but only one dedup row per table, so
counting is `SELECT count(*) WHERE bucket = ?` rather than `COUNT(DISTINCT user_id)` over
a growing raw table — see [Database Design](#database-design) for why.

**Two dedup tables instead of one**: monthly-unique cannot be derived by summing daily
counts (a user active on 3 days in a month is 1 monthly-unique user, not 3), so a separate
monthly dedup table is maintained directly rather than computed from the daily one at query
time.

**Idempotency**: `RecordLogin` is safe to retry — the same `(user_id, login_at)` recorded
twice still only counts once, because the dedup insert is `ON CONFLICT DO NOTHING`.

## Database design

```sql
CREATE TABLE user_logins (
    id         BIGSERIAL PRIMARY KEY,
    user_id    UUID NOT NULL,
    login_at   TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_user_logins_login_at ON user_logins (login_at);
CREATE INDEX idx_user_logins_user_id  ON user_logins (user_id);

CREATE TABLE daily_active_users (
    day     DATE NOT NULL,
    user_id UUID NOT NULL,
    PRIMARY KEY (day, user_id)
);

CREATE TABLE monthly_active_users (
    month   DATE NOT NULL,   -- first day of the month
    user_id UUID NOT NULL,
    PRIMARY KEY (month, user_id)
);
CREATE INDEX idx_dau_day   ON daily_active_users   (day);
CREATE INDEX idx_mau_month ON monthly_active_users (month);
```

Full file: [migrations/001_init.sql](migrations/001_init.sql).

- `user_logins` is the append-only source of truth — every login is recorded here even if
  it doesn't change a unique count, which matters for audit/replay and for rebuilding the
  dedup tables if their logic ever changes.
- The dedup tables' composite primary key `(bucket, user_id)` is what actually enforces
  uniqueness — it's a database constraint, not application logic, so it holds even under
  concurrent writes for the same user.
- Query cost: `count(*) FROM daily_active_users WHERE day = ?` is an index-range scan on
  `idx_dau_day` over rows already deduplicated — it doesn't re-scan or re-distinct the raw
  log, so it stays cheap as `user_logins` grows into the billions of rows.
- Ingest cost: 3 writes per login instead of 1, but all in a single transaction, and two of
  the three are no-ops (`ON CONFLICT DO NOTHING`) after a user's first login in a given
  bucket — a reasonable trade for making reads cheap, since reads (dashboards, reports) are
  typically far more frequent than writes here.
- At larger scale, `user_logins` would be partitioned by month (Postgres native range
  partitioning on `login_at`) to keep the raw log's indexes small and to make old partitions
  droppable/archivable independently.

## Architecture

```
cmd/server           entrypoint: config, DB connection, HTTP server wiring
internal/store        Store interface + Postgres implementation + in-memory
                       implementation (used by tests, no DB needed)
internal/analytics     business logic: parses date/month inputs, calls Store
internal/api           HTTP handlers, thin translation to/from JSON
migrations             SQL schema
testdata               sample data with expected query outputs
```

The service is stateless (all state lives in Postgres), so it can run as multiple
replicas behind a load balancer. `internal/store.Store` is an interface specifically so
`internal/analytics` and `internal/api` can be unit-tested without a real database.
