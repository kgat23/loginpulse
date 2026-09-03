-- Raw event log: source of truth for every login.
CREATE TABLE user_logins (
    id         BIGSERIAL PRIMARY KEY,
    user_id    UUID NOT NULL,
    login_at   TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_user_logins_login_at ON user_logins (login_at);
CREATE INDEX idx_user_logins_user_id ON user_logins (user_id);

-- Dedup tables, one row per (bucket, user), maintained on ingest.
-- Bucketing is always done in UTC.
CREATE TABLE daily_active_users (
    day     DATE NOT NULL,
    user_id UUID NOT NULL,
    PRIMARY KEY (day, user_id)
);

CREATE TABLE monthly_active_users (
    month   DATE NOT NULL, -- first day of the month, e.g. 2026-09-01
    user_id UUID NOT NULL,
    PRIMARY KEY (month, user_id)
);

CREATE INDEX idx_dau_day ON daily_active_users (day);
CREATE INDEX idx_mau_month ON monthly_active_users (month);
