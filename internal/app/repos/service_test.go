package repos_test

import (
	"context"
	"errors"
	"testing"

	"github.com/tbxark/gh-stars/internal/app/repos"
	"github.com/tbxark/gh-stars/internal/domain"
	"github.com/tbxark/gh-stars/internal/domain/testdata"
	"github.com/tbxark/gh-stars/internal/github"
	"github.com/tbxark/gh-stars/internal/testutil"
)

func TestService_LoadDetails_Success(t *testing.T) {
	mockClient := github.NewMockClient()
	expectedDetails := testdata.SampleRepoDetails()
	mockClient.GetRepoDetailsFunc = func(ctx context.Context, fullName, token string) (domain.RepoDetails, error) {
		return expectedDetails, nil
	}

	service := repos.Service{GH: mockClient}
	ctx := context.Background()

	details, err := service.LoadDetails(ctx, "golang/go", "token123")

	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, expectedDetails.FullName, details.FullName)
	testutil.AssertEqual(t, 1, mockClient.CallCounts.GetRepoDetails)
}

func TestService_LoadDetails_EmptyFullName(t *testing.T) {
	mockClient := github.NewMockClient()
	mockClient.GetRepoDetailsFunc = func(ctx context.Context, fullName, token string) (domain.RepoDetails, error) {
		return domain.RepoDetails{}, errors.New("repo full name is required")
	}

	service := repos.Service{GH: mockClient}
	ctx := context.Background()

	details, err := service.LoadDetails(ctx, "", "token123")

	testutil.AssertError(t, err)
	testutil.AssertEqual(t, "", details.FullName)
}

func TestService_LoadDetails_ClientError(t *testing.T) {
	mockClient := github.NewMockClient()
	expectedError := errors.New("404 not found")
	mockClient.GetRepoDetailsFunc = func(ctx context.Context, fullName, token string) (domain.RepoDetails, error) {
		return domain.RepoDetails{}, expectedError
	}

	service := repos.Service{GH: mockClient}
	ctx := context.Background()

	details, err := service.LoadDetails(ctx, "nonexistent/repo", "token123")

	testutil.AssertError(t, err)
	testutil.AssertEqual(t, "", details.FullName)
	testutil.AssertEqual(t, 1, mockClient.CallCounts.GetRepoDetails)
}

func TestService_LoadDetails_ContextCancellation(t *testing.T) {
	mockClient := github.NewMockClient()
	mockClient.GetRepoDetailsFunc = func(ctx context.Context, fullName, token string) (domain.RepoDetails, error) {
		return domain.RepoDetails{}, context.Canceled
	}

	service := repos.Service{GH: mockClient}
	ctx, cancel := testutil.WithCancel()
	cancel()

	details, err := service.LoadDetails(ctx, "golang/go", "token123")

	testutil.AssertError(t, err)
	testutil.AssertEqual(t, "", details.FullName)
}
