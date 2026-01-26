package stars

import (
	"context"
	"fmt"
	"sync"

	"github.com/tbxark/gh-stars/internal/domain"
)

// MockService is a mock implementation of stars.Loader for testing
type MockService struct {
	// LoadStarredFunc allows overriding the behavior in tests
	LoadStarredFunc func(ctx context.Context, username, token string, perPage int) ([]domain.Repo, error)

	// CallCounts tracks how many times each method was called
	CallCounts struct {
		mu          sync.Mutex
		LoadStarred int
	}
}

var _ Loader = (*MockService)(nil) // Compile-time interface check

// NewMockService creates a new mock service with default behavior
func NewMockService() *MockService {
	return &MockService{
		LoadStarredFunc: func(ctx context.Context, username, token string, perPage int) ([]domain.Repo, error) {
			return nil, fmt.Errorf("mock LoadStarred not implemented")
		},
	}
}

// LoadStarred implements the Loader interface
func (m *MockService) LoadStarred(ctx context.Context, username, token string, perPage int) ([]domain.Repo, error) {
	m.CallCounts.mu.Lock()
	m.CallCounts.LoadStarred++
	m.CallCounts.mu.Unlock()
	return m.LoadStarredFunc(ctx, username, token, perPage)
}

// ResetCallCounts resets all call counters (useful between test cases)
func (m *MockService) ResetCallCounts() {
	m.CallCounts.mu.Lock()
	m.CallCounts.LoadStarred = 0
	m.CallCounts.mu.Unlock()
}

// GetLoadStarredCount returns the current call count in a thread-safe manner
func (m *MockService) GetLoadStarredCount() int {
	m.CallCounts.mu.Lock()
	defer m.CallCounts.mu.Unlock()
	return m.CallCounts.LoadStarred
}
