package github

import (
	"context"
	"fmt"

	"github.com/TBXark/gh-stars/internal/domain"
)

// MockClient is a mock implementation of GitHub client for testing
type MockClient struct {
	// ListStarredFunc allows overriding the behavior in tests
	ListStarredFunc func(ctx context.Context, username, token string, perPage int) ([]domain.Repo, error)

	// GetRepoDetailsFunc allows overriding the behavior in tests
	GetRepoDetailsFunc func(ctx context.Context, fullName, token string) (domain.RepoDetails, error)

	// CallCounts tracks how many times each method was called
	CallCounts struct {
		ListStarred    int
		GetRepoDetails int
	}
}

// NewMockClient creates a new mock client with default behavior
func NewMockClient() *MockClient {
	return &MockClient{
		ListStarredFunc: func(ctx context.Context, username, token string, perPage int) ([]domain.Repo, error) {
			return nil, fmt.Errorf("mock ListStarred not implemented")
		},
		GetRepoDetailsFunc: func(ctx context.Context, fullName, token string) (domain.RepoDetails, error) {
			return domain.RepoDetails{}, fmt.Errorf("mock GetRepoDetails not implemented")
		},
	}
}

// ListStarred implements the Client interface
func (m *MockClient) ListStarred(ctx context.Context, username, token string, perPage int) ([]domain.Repo, error) {
	m.CallCounts.ListStarred++
	return m.ListStarredFunc(ctx, username, token, perPage)
}

// GetRepoDetails implements the Client interface
func (m *MockClient) GetRepoDetails(ctx context.Context, fullName, token string) (domain.RepoDetails, error) {
	m.CallCounts.GetRepoDetails++
	return m.GetRepoDetailsFunc(ctx, fullName, token)
}

// ResetCallCounts resets all call counters (useful between test cases)
func (m *MockClient) ResetCallCounts() {
	m.CallCounts.ListStarred = 0
	m.CallCounts.GetRepoDetails = 0
}
