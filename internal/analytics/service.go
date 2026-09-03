// Package analytics implements the business logic for ingesting logins and
// answering unique-user queries.
package analytics

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"loginpulse/internal/store"
)

const (
	dayLayout   = "2006-01-02"
	monthLayout = "2006-01"
)

type Service struct {
	store store.Store
}

func NewService(s store.Store) *Service {
	return &Service{store: s}
}

// RecordLogin ingests a login event for userID at loginAt.
func (s *Service) RecordLogin(ctx context.Context, userID uuid.UUID, loginAt time.Time) error {
	return s.store.RecordLogin(ctx, userID, loginAt)
}

// GetDailyUniqueUsers returns the unique login count for the UTC day given
// as "YYYY-MM-DD".
func (s *Service) GetDailyUniqueUsers(ctx context.Context, date string) (int, error) {
	t, err := time.Parse(dayLayout, date)
	if err != nil {
		return 0, fmt.Errorf("invalid date %q, want YYYY-MM-DD: %w", date, err)
	}
	return s.store.DailyUniqueUsers(ctx, t)
}

// GetMonthlyUniqueUsers returns the unique login count for the UTC month
// given as "YYYY-MM".
func (s *Service) GetMonthlyUniqueUsers(ctx context.Context, month string) (int, error) {
	t, err := time.Parse(monthLayout, month)
	if err != nil {
		return 0, fmt.Errorf("invalid month %q, want YYYY-MM: %w", month, err)
	}
	return s.store.MonthlyUniqueUsers(ctx, t)
}
