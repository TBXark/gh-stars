package stars_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
	"github.com/TBXark/gh-stars/internal/app/stars"
	"github.com/TBXark/gh-stars/internal/domain"
	"github.com/TBXark/gh-stars/internal/domain/testdata"
	"github.com/TBXark/gh-stars/internal/testutil"
	uistars "github.com/TBXark/gh-stars/internal/ui/stars"
)

func TestVM_ConcurrentLoad_NoRaceCondition(t *testing.T) {
	// Initialize test Fyne app
	_ = test.NewApp()

	mockSvc := stars.NewMockService()
	mockSvc.LoadStarredFunc = func(ctx context.Context, username, token string, perPage int) ([]domain.Repo, error) {
		time.Sleep(50 * time.Millisecond)
		return testdata.SampleRepoList(), nil
	}

	runOnMain := func(f func()) { f() }
	vm := uistars.NewVM(mockSvc, runOnMain)
	_ = vm.Username.Set("testuser")
	_ = vm.Token.Set("token123")

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			vm.Load()
		}()
	}

	wg.Wait()
	time.Sleep(200 * time.Millisecond)

	loading, _ := vm.Loading.Get()
	testutil.AssertFalse(t, loading, "loading should be false after operations")
}

func TestVM_ContextCancellation_ConcurrentOperations(t *testing.T) {
	callCount := 0
	var mu sync.Mutex

	mockSvc := stars.NewMockService()
	mockSvc.LoadStarredFunc = func(ctx context.Context, username, token string, perPage int) ([]domain.Repo, error) {
		mu.Lock()
		callCount++
		mu.Unlock()

		select {
		case <-time.After(500 * time.Millisecond):
			return testdata.SampleRepoList(), nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	runOnMain := func(f func()) { f() }
	vm := uistars.NewVM(mockSvc, runOnMain)
	_ = vm.Username.Set("testuser")
	_ = vm.Token.Set("token123")

	vm.Load()
	time.Sleep(50 * time.Millisecond)

	vm.Load()
	time.Sleep(50 * time.Millisecond)

	vm.Load()
	time.Sleep(600 * time.Millisecond)

	mu.Lock()
	actualCallCount := callCount
	mu.Unlock()

	testutil.AssertTrue(t, actualCallCount >= 3, "should have called service at least 3 times")

	status, _ := vm.Status.Get()
	testutil.AssertEqual(t, "Loaded", status)
}
