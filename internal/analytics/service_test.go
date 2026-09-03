package analytics

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"loginpulse/internal/store"
)

func newTestService() *Service {
	return NewService(store.NewMemoryStore())
}

func mustParse(t *testing.T, layout, value string) time.Time {
	t.Helper()
	tm, err := time.Parse(layout, value)
	if err != nil {
		t.Fatalf("parse %q: %v", value, err)
	}
	return tm
}

func TestDuplicateLoginsSameDayCountOnce(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	user := uuid.New()

	for _, ts := range []string{
		"2026-09-02T08:00:00Z",
		"2026-09-02T09:30:00Z",
		"2026-09-02T23:59:59Z",
	} {
		if err := svc.RecordLogin(ctx, user, mustParse(t, time.RFC3339, ts)); err != nil {
			t.Fatalf("RecordLogin: %v", err)
		}
	}

	got, err := svc.GetDailyUniqueUsers(ctx, "2026-09-02")
	if err != nil {
		t.Fatalf("GetDailyUniqueUsers: %v", err)
	}
	if got != 1 {
		t.Errorf("daily unique users = %d, want 1", got)
	}
}

func TestLoginsAcrossDayBoundary(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	userA, userB := uuid.New(), uuid.New()

	svc.RecordLogin(ctx, userA, mustParse(t, time.RFC3339, "2026-09-02T23:59:59Z"))
	svc.RecordLogin(ctx, userB, mustParse(t, time.RFC3339, "2026-09-03T00:00:01Z"))

	day2, _ := svc.GetDailyUniqueUsers(ctx, "2026-09-02")
	day3, _ := svc.GetDailyUniqueUsers(ctx, "2026-09-03")

	if day2 != 1 {
		t.Errorf("2026-09-02 daily unique = %d, want 1", day2)
	}
	if day3 != 1 {
		t.Errorf("2026-09-03 daily unique = %d, want 1", day3)
	}
}

func TestLoginsAcrossMonthBoundary(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	user := uuid.New()

	svc.RecordLogin(ctx, user, mustParse(t, time.RFC3339, "2026-08-31T23:59:59Z"))
	svc.RecordLogin(ctx, user, mustParse(t, time.RFC3339, "2026-09-01T00:00:01Z"))

	aug, _ := svc.GetMonthlyUniqueUsers(ctx, "2026-08")
	sep, _ := svc.GetMonthlyUniqueUsers(ctx, "2026-09")

	if aug != 1 {
		t.Errorf("2026-08 monthly unique = %d, want 1", aug)
	}
	if sep != 1 {
		t.Errorf("2026-09 monthly unique = %d, want 1", sep)
	}
}

func TestMonthlyUniqueDoesNotDoubleCountAcrossDays(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	user := uuid.New()

	svc.RecordLogin(ctx, user, mustParse(t, time.RFC3339, "2026-09-01T00:00:00Z"))
	svc.RecordLogin(ctx, user, mustParse(t, time.RFC3339, "2026-09-15T12:00:00Z"))
	svc.RecordLogin(ctx, user, mustParse(t, time.RFC3339, "2026-09-30T23:59:59Z"))

	got, err := svc.GetMonthlyUniqueUsers(ctx, "2026-09")
	if err != nil {
		t.Fatalf("GetMonthlyUniqueUsers: %v", err)
	}
	if got != 1 {
		t.Errorf("monthly unique users = %d, want 1 (same user, multiple days)", got)
	}
}

func TestNonUTCTimestampNormalizedToUTCDay(t *testing.T) {
	// 2026-09-02 23:30 in UTC+2 is 2026-09-02 21:30 UTC, still the same
	// UTC day. 2026-09-02 23:30 in UTC-5 is 2026-09-03 04:30 UTC, the next
	// UTC day. Both must bucket by UTC day, not by the offset in the
	// timestamp.
	svc := newTestService()
	ctx := context.Background()
	userPlus2, userMinus5 := uuid.New(), uuid.New()

	plus2 := mustParse(t, time.RFC3339, "2026-09-02T23:30:00+02:00")
	minus5 := mustParse(t, time.RFC3339, "2026-09-02T23:30:00-05:00")

	svc.RecordLogin(ctx, userPlus2, plus2)
	svc.RecordLogin(ctx, userMinus5, minus5)

	sep2, _ := svc.GetDailyUniqueUsers(ctx, "2026-09-02")
	sep3, _ := svc.GetDailyUniqueUsers(ctx, "2026-09-03")

	if sep2 != 1 {
		t.Errorf("2026-09-02 daily unique = %d, want 1 (the +02:00 login)", sep2)
	}
	if sep3 != 1 {
		t.Errorf("2026-09-03 daily unique = %d, want 1 (the -05:00 login)", sep3)
	}
}

func TestMultipleDistinctUsersCounted(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		svc.RecordLogin(ctx, uuid.New(), mustParse(t, time.RFC3339, "2026-09-02T10:00:00Z"))
	}

	got, err := svc.GetDailyUniqueUsers(ctx, "2026-09-02")
	if err != nil {
		t.Fatalf("GetDailyUniqueUsers: %v", err)
	}
	if got != 5 {
		t.Errorf("daily unique users = %d, want 5", got)
	}
}

func TestInvalidDateFormat(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	if _, err := svc.GetDailyUniqueUsers(ctx, "09/02/2026"); err == nil {
		t.Error("expected error for malformed date, got nil")
	}
	if _, err := svc.GetMonthlyUniqueUsers(ctx, "2026-9"); err == nil {
		t.Error("expected error for malformed month, got nil")
	}
}

func TestNoLogins(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	got, err := svc.GetDailyUniqueUsers(ctx, "2026-01-01")
	if err != nil {
		t.Fatalf("GetDailyUniqueUsers: %v", err)
	}
	if got != 0 {
		t.Errorf("daily unique users = %d, want 0", got)
	}
}
