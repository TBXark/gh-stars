package repos

import (
	"context"
	"fmt"
	"sync"

	"github.com/tbxark/gh-stars/internal/domain"
)

// MockService is a mock implementation of repos.Loader for testing
type MockService struct {
	// LoadDetailsFunc allows overriding the behavior in tests
	LoadDetailsFunc func(ctx context.Context, fullName, token string) (domain.RepoDetails, error)

	// CallCounts tracks how many times each method was called
	CallCounts struct {
		mu          sync.Mutex
		LoadDetails int
	}
}

var _ Loader = (*MockService)(nil) // Compile-time interface check

// NewMockService creates a new mock service with default behavior
func NewMockService() *MockService {
	return &MockService{
		LoadDetailsFunc: func(ctx context.Context, fullName, token string) (domain.RepoDetails, error) {
			return domain.RepoDetails{}, fmt.Errorf("mock LoadDetails not implemented")
		},
	}
}

// LoadDetails implements the Loader interface
func (m *MockService) LoadDetails(ctx context.Context, fullName, token string) (domain.RepoDetails, error) {
	m.CallCounts.mu.Lock()
	m.CallCounts.LoadDetails++
	m.CallCounts.mu.Unlock()
	return m.LoadDetailsFunc(ctx, fullName, token)
}

// ResetCallCounts resets all call counters (useful between test cases)
func (m *MockService) ResetCallCounts() {
	m.CallCounts.mu.Lock()
	m.CallCounts.LoadDetails = 0
	m.CallCounts.mu.Unlock()
}

// GetLoadDetailsCount returns the current call count in a thread-safe manner
func (m *MockService) GetLoadDetailsCount() int {
	m.CallCounts.mu.Lock()
	defer m.CallCounts.mu.Unlock()
	return m.CallCounts.LoadDetails
}
