// Package store defines the persistence interface used by the analytics
// service, decoupled from any specific database driver.
package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Store records login events and answers uniqueness queries against them.
type Store interface {
	// RecordLogin ingests a single login event. It is idempotent: the same
	// (user, timestamp) pair recorded twice only counts once toward the
	// daily/monthly unique totals.
	RecordLogin(ctx context.Context, userID uuid.UUID, loginAt time.Time) error

	// DailyUniqueUsers returns the number of distinct users active on the
	// UTC calendar day containing day.
	DailyUniqueUsers(ctx context.Context, day time.Time) (int, error)

	// MonthlyUniqueUsers returns the number of distinct users active in the
	// UTC calendar month containing month.
	MonthlyUniqueUsers(ctx context.Context, month time.Time) (int, error)
}
