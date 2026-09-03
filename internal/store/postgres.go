package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PostgresStore is a Store backed by PostgreSQL.
type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) RecordLogin(ctx context.Context, userID uuid.UUID, loginAt time.Time) error {
	loginAt = loginAt.UTC()
	day := truncDay(loginAt)
	month := truncMonth(loginAt)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO user_logins (user_id, login_at) VALUES ($1, $2)`,
		userID, loginAt); err != nil {
		return fmt.Errorf("insert user_logins: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO daily_active_users (day, user_id) VALUES ($1, $2)
		 ON CONFLICT DO NOTHING`,
		day, userID); err != nil {
		return fmt.Errorf("insert daily_active_users: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO monthly_active_users (month, user_id) VALUES ($1, $2)
		 ON CONFLICT DO NOTHING`,
		month, userID); err != nil {
		return fmt.Errorf("insert monthly_active_users: %w", err)
	}

	return tx.Commit()
}

func (s *PostgresStore) DailyUniqueUsers(ctx context.Context, day time.Time) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT count(*) FROM daily_active_users WHERE day = $1`,
		truncDay(day.UTC())).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("query daily_active_users: %w", err)
	}
	return count, nil
}

func (s *PostgresStore) MonthlyUniqueUsers(ctx context.Context, month time.Time) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT count(*) FROM monthly_active_users WHERE month = $1`,
		truncMonth(month.UTC())).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("query monthly_active_users: %w", err)
	}
	return count, nil
}

func truncDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func truncMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}
