package stars_test

import (
	"context"
	"errors"
	"testing"

	"github.com/TBXark/gh-stars/internal/app/stars"
	"github.com/TBXark/gh-stars/internal/domain"
	"github.com/TBXark/gh-stars/internal/domain/testdata"
	"github.com/TBXark/gh-stars/internal/github"
	"github.com/TBXark/gh-stars/internal/testutil"
)

func TestService_LoadStarred_Success(t *testing.T) {
	// Arrange
	mockClient := github.NewMockClient()
	expectedRepos := testdata.SampleRepoList()
	mockClient.ListStarredFunc = func(ctx context.Context, username, token string, perPage int) ([]domain.Repo, error) {
		return expectedRepos, nil
	}

	service := stars.Service{GH: mockClient}
	ctx := context.Background()

	// Act
	repos, err := service.LoadStarred(ctx, "testuser", "token123", 30)

	// Assert
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, len(expectedRepos), len(repos))
	testutil.AssertEqual(t, 1, mockClient.CallCounts.ListStarred)
}

func TestService_LoadStarred_EmptyUsername(t *testing.T) {
	// Arrange
	mockClient := github.NewMockClient()
	mockClient.ListStarredFunc = func(ctx context.Context, username, token string, perPage int) ([]domain.Repo, error) {
		return nil, errors.New("username is required")
	}

	service := stars.Service{GH: mockClient}
	ctx := context.Background()

	// Act
	repos, err := service.LoadStarred(ctx, "", "token123", 30)

	// Assert
	testutil.AssertError(t, err)
	testutil.AssertEqual(t, 0, len(repos))
}

func TestService_LoadStarred_ClientError(t *testing.T) {
	// Arrange
	mockClient := github.NewMockClient()
	expectedError := errors.New("API rate limit exceeded")
	mockClient.ListStarredFunc = func(ctx context.Context, username, token string, perPage int) ([]domain.Repo, error) {
		return nil, expectedError
	}

	service := stars.Service{GH: mockClient}
	ctx := context.Background()

	// Act
	repos, err := service.LoadStarred(ctx, "testuser", "token123", 30)

	// Assert
	testutil.AssertError(t, err)
	testutil.AssertEqual(t, 0, len(repos))
	testutil.AssertEqual(t, 1, mockClient.CallCounts.ListStarred)
}

func TestService_LoadStarred_ContextCancellation(t *testing.T) {
	// Arrange
	mockClient := github.NewMockClient()
	mockClient.ListStarredFunc = func(ctx context.Context, username, token string, perPage int) ([]domain.Repo, error) {
		return nil, context.Canceled
	}

	service := stars.Service{GH: mockClient}
	ctx, cancel := testutil.WithCancel()
	cancel() // Cancel immediately

	// Act
	repos, err := service.LoadStarred(ctx, "testuser", "token123", 30)

	// Assert
	testutil.AssertError(t, err)
	testutil.AssertEqual(t, 0, len(repos))
}

func TestService_LoadStarred_EmptyResult(t *testing.T) {
	// Arrange
	mockClient := github.NewMockClient()
	mockClient.ListStarredFunc = func(ctx context.Context, username, token string, perPage int) ([]domain.Repo, error) {
		return []domain.Repo{}, nil
	}

	service := stars.Service{GH: mockClient}
	ctx := context.Background()

	// Act
	repos, err := service.LoadStarred(ctx, "userwithnorepos", "token123", 30)

	// Assert
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, 0, len(repos))
	testutil.AssertEqual(t, 1, mockClient.CallCounts.ListStarred)
}

func TestService_LoadStarred_ValidatesPerPage(t *testing.T) {
	// Arrange
	mockClient := github.NewMockClient()
	var receivedPerPage int
	mockClient.ListStarredFunc = func(ctx context.Context, username, token string, perPage int) ([]domain.Repo, error) {
		receivedPerPage = perPage
		return testdata.SampleRepoList(), nil
	}

	service := stars.Service{GH: mockClient}
	ctx := context.Background()

	// Act
	_, err := service.LoadStarred(ctx, "testuser", "token123", 50)

	// Assert
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, 50, receivedPerPage)
}
