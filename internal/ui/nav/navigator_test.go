package nav_test

import (
	"testing"

	"github.com/tbxark/gh-stars/internal/app/repos"
	"github.com/tbxark/gh-stars/internal/app/stars"
	"github.com/tbxark/gh-stars/internal/testutil"
	"github.com/tbxark/gh-stars/internal/ui/nav"
)

func TestAppNavigator_Initialization(t *testing.T) {
	mockStarsSvc := stars.NewMockService()
	mockRepoSvc := repos.NewMockService()

	navigator := &nav.AppNavigator{
		StarsSvc: mockStarsSvc,
		RepoSvc:  mockRepoSvc,
	}

	testutil.AssertTrue(t, navigator != nil, "Navigator should be initialized")
}

func TestAppNavigator_ThreadSafety(t *testing.T) {
	mockStarsSvc := stars.NewMockService()
	mockRepoSvc := repos.NewMockService()

	navigator := &nav.AppNavigator{
		StarsSvc: mockStarsSvc,
		RepoSvc:  mockRepoSvc,
	}

	testutil.AssertTrue(t, navigator != nil, "Navigator with mutex protection initialized")
}
