package greeter

import (
	"context"
	"sync"
)

// Repository is the data layer of the greeter service.
// Swap InMemoryRepository for a real database-backed implementation.
type Repository interface {
	RecordGreeting(ctx context.Context, name string) error
	CountGreetings(ctx context.Context, name string) (int, error)
}

type InMemoryRepository struct {
	mu     sync.Mutex
	counts map[string]int
}

func NewRepository() *InMemoryRepository {
	return &InMemoryRepository{counts: make(map[string]int)}
}

func (r *InMemoryRepository) RecordGreeting(_ context.Context, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.counts[name]++
	return nil
}

func (r *InMemoryRepository) CountGreetings(_ context.Context, name string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.counts[name], nil
}
