package store

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MemoryStore is an in-memory Store used for unit tests. It mirrors the
// bucketing/dedup semantics of PostgresStore without needing a database.
type MemoryStore struct {
	mu      sync.Mutex
	daily   map[time.Time]map[uuid.UUID]struct{}
	monthly map[time.Time]map[uuid.UUID]struct{}
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		daily:   make(map[time.Time]map[uuid.UUID]struct{}),
		monthly: make(map[time.Time]map[uuid.UUID]struct{}),
	}
}

func (s *MemoryStore) RecordLogin(_ context.Context, userID uuid.UUID, loginAt time.Time) error {
	loginAt = loginAt.UTC()
	day := truncDay(loginAt)
	month := truncMonth(loginAt)

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.daily[day] == nil {
		s.daily[day] = make(map[uuid.UUID]struct{})
	}
	s.daily[day][userID] = struct{}{}

	if s.monthly[month] == nil {
		s.monthly[month] = make(map[uuid.UUID]struct{})
	}
	s.monthly[month][userID] = struct{}{}

	return nil
}

func (s *MemoryStore) DailyUniqueUsers(_ context.Context, day time.Time) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.daily[truncDay(day.UTC())]), nil
}

func (s *MemoryStore) MonthlyUniqueUsers(_ context.Context, month time.Time) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.monthly[truncMonth(month.UTC())]), nil
}
