package stars_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/TBXark/gh-stars/internal/app/stars"
	"github.com/TBXark/gh-stars/internal/domain"
	"github.com/TBXark/gh-stars/internal/domain/testdata"
	"github.com/TBXark/gh-stars/internal/testutil"
	uistars "github.com/TBXark/gh-stars/internal/ui/stars"
)

func TestVM_Load_Success(t *testing.T) {
	mockSvc := stars.NewMockService()
	expectedRepos := testdata.SampleRepoList()
	mockSvc.LoadStarredFunc = func(ctx context.Context, username, token string, perPage int) ([]domain.Repo, error) {
		return expectedRepos, nil
	}

	var runOnMainCalled bool
	runOnMain := func(f func()) {
		runOnMainCalled = true
		f()
	}

	vm := uistars.NewVM(mockSvc, runOnMain)
	_ = vm.Username.Set("testuser")
	_ = vm.Token.Set("token123")
	_ = vm.PerPage.Set("50")

	vm.Load()
	time.Sleep(50 * time.Millisecond)

	status, _ := vm.Status.Get()
	loading, _ := vm.Loading.Get()
	errorMsg, _ := vm.Error.Get()

	testutil.AssertTrue(t, runOnMainCalled, "runOnMain should be called")
	testutil.AssertEqual(t, "Loaded", status)
	testutil.AssertFalse(t, loading, "loading should be false after completion")
	testutil.AssertEqual(t, "", errorMsg)
	testutil.AssertEqual(t, 1, mockSvc.GetLoadStarredCount())
}

func TestVM_Load_ServiceError(t *testing.T) {
	mockSvc := stars.NewMockService()
	expectedError := errors.New("API rate limit exceeded")
	mockSvc.LoadStarredFunc = func(ctx context.Context, username, token string, perPage int) ([]domain.Repo, error) {
		return nil, expectedError
	}

	runOnMain := func(f func()) { f() }

	vm := uistars.NewVM(mockSvc, runOnMain)
	_ = vm.Username.Set("testuser")
	_ = vm.Token.Set("token123")

	vm.Load()
	time.Sleep(50 * time.Millisecond)

	status, _ := vm.Status.Get()
	loading, _ := vm.Loading.Get()
	errorMsg, _ := vm.Error.Get()

	testutil.AssertEqual(t, "Load failed", status)
	testutil.AssertFalse(t, loading, "loading should be false after error")
	testutil.AssertEqual(t, expectedError.Error(), errorMsg)
}

func TestVM_Load_InvalidPerPage(t *testing.T) {
	mockSvc := stars.NewMockService()
	runOnMain := func(f func()) { f() }

	vm := uistars.NewVM(mockSvc, runOnMain)
	_ = vm.Username.Set("testuser")
	_ = vm.Token.Set("token123")
	_ = vm.PerPage.Set("invalid")

	vm.Load()
	time.Sleep(50 * time.Millisecond)

	status, _ := vm.Status.Get()
	errorMsg, _ := vm.Error.Get()

	testutil.AssertEqual(t, "Load failed", status)
	testutil.AssertEqual(t, "per page must be a number", errorMsg)
	testutil.AssertEqual(t, 0, mockSvc.GetLoadStarredCount())
}

func TestVM_Clear(t *testing.T) {
	mockSvc := stars.NewMockService()
	runOnMain := func(f func()) { f() }

	vm := uistars.NewVM(mockSvc, runOnMain)

	vm.Clear()

	status, _ := vm.Status.Get()
	errorMsg, _ := vm.Error.Get()

	testutil.AssertEqual(t, "Cleared", status)
	testutil.AssertEqual(t, "", errorMsg)
}

func TestVM_RepoAt_ValidIndex(t *testing.T) {
	mockSvc := stars.NewMockService()
	expectedRepos := testdata.SampleRepoList()
	mockSvc.LoadStarredFunc = func(ctx context.Context, username, token string, perPage int) ([]domain.Repo, error) {
		return expectedRepos, nil
	}

	runOnMain := func(f func()) { f() }
	vm := uistars.NewVM(mockSvc, runOnMain)
	_ = vm.Username.Set("testuser")

	vm.Load()
	time.Sleep(50 * time.Millisecond)

	repo, ok := vm.RepoAt(0)

	testutil.AssertTrue(t, ok, "should find repo at index 0")
	testutil.AssertEqual(t, expectedRepos[0].FullName, repo.FullName)
}

func TestVM_RepoAt_InvalidIndex(t *testing.T) {
	mockSvc := stars.NewMockService()
	runOnMain := func(f func()) { f() }

	vm := uistars.NewVM(mockSvc, runOnMain)

	_, ok := vm.RepoAt(-1)
	testutil.AssertFalse(t, ok, "should not find repo at negative index")

	_, ok = vm.RepoAt(999)
	testutil.AssertFalse(t, ok, "should not find repo at out-of-bounds index")
}

func TestVM_DefaultValues(t *testing.T) {
	mockSvc := stars.NewMockService()
	runOnMain := func(f func()) { f() }

	vm := uistars.NewVM(mockSvc, runOnMain)

	perPage, _ := vm.PerPage.Get()
	status, _ := vm.Status.Get()
	loading, _ := vm.Loading.Get()

	testutil.AssertEqual(t, "100", perPage)
	testutil.AssertEqual(t, "Ready", status)
	testutil.AssertFalse(t, loading, "loading should be false by default")
}
