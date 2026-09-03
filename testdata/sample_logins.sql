-- Sample data: 3 users, spanning a day boundary (Sep 2 -> Sep 3) and a
-- month boundary (Aug 31 -> Sep 1). Run against a fresh database to see
-- non-trivial daily/monthly counts.
INSERT INTO user_logins (user_id, login_at) VALUES
    ('11111111-1111-1111-1111-111111111111', '2026-08-31 23:00:00+00'),
    ('22222222-2222-2222-2222-222222222222', '2026-08-31 23:30:00+00'),
    ('11111111-1111-1111-1111-111111111111', '2026-09-01 00:15:00+00'),
    ('33333333-3333-3333-3333-333333333333', '2026-09-02 08:00:00+00'),
    ('11111111-1111-1111-1111-111111111111', '2026-09-02 09:00:00+00'),
    ('11111111-1111-1111-1111-111111111111', '2026-09-02 20:00:00+00'), -- duplicate day, same user
    ('22222222-2222-2222-2222-222222222222', '2026-09-03 00:05:00+00');

INSERT INTO daily_active_users (day, user_id)
SELECT DISTINCT (login_at AT TIME ZONE 'UTC')::date, user_id FROM user_logins
ON CONFLICT DO NOTHING;

INSERT INTO monthly_active_users (month, user_id)
SELECT DISTINCT date_trunc('month', login_at AT TIME ZONE 'UTC')::date, user_id FROM user_logins
ON CONFLICT DO NOTHING;

-- Expected:
--   GetDailyUniqueUsers("2026-08-31")   -> 2  (user 1, user 2)
--   GetDailyUniqueUsers("2026-09-02")   -> 2  (user 1 twice + user 3, dedup'd)
--   GetMonthlyUniqueUsers("2026-08")    -> 2  (user 1, user 2)
--   GetMonthlyUniqueUsers("2026-09")    -> 3  (user 1, user 2, user 3)
