package details

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2/data/binding"

	"github.com/TBXark/gh-stars/internal/app/repos"
	"github.com/TBXark/gh-stars/internal/domain"
)

type VM struct {
	FullName string
	Token    string

	Loading binding.Bool
	Status  binding.String
	Error   binding.String

	Name          binding.String
	Description   binding.String
	Language      binding.String
	Homepage      binding.String
	DefaultBranch binding.String
	License       binding.String
	Topics        binding.String
	Stars         binding.String
	Forks         binding.String
	Watchers      binding.String
	OpenIssues    binding.String
	Size          binding.String
	UpdatedAt     binding.String
	CreatedAt     binding.String
	PushedAt      binding.String
	Private       binding.String
	HTMLURL       binding.String

	svc       repos.Service
	runOnMain func(func())

	mu     sync.Mutex
	cancel context.CancelFunc
}

func NewVM(svc repos.Service, fullName, token string, runOnMain func(func())) *VM {
	vm := &VM{
		FullName:      fullName,
		Token:         token,
		Loading:       binding.NewBool(),
		Status:        binding.NewString(),
		Error:         binding.NewString(),
		Name:          binding.NewString(),
		Description:   binding.NewString(),
		Language:      binding.NewString(),
		Homepage:      binding.NewString(),
		DefaultBranch: binding.NewString(),
		License:       binding.NewString(),
		Topics:        binding.NewString(),
		Stars:         binding.NewString(),
		Forks:         binding.NewString(),
		Watchers:      binding.NewString(),
		OpenIssues:    binding.NewString(),
		Size:          binding.NewString(),
		UpdatedAt:     binding.NewString(),
		CreatedAt:     binding.NewString(),
		PushedAt:      binding.NewString(),
		Private:       binding.NewString(),
		HTMLURL:       binding.NewString(),
		svc:           svc,
		runOnMain:     runOnMain,
	}
	if vm.runOnMain == nil {
		vm.runOnMain = func(f func()) { f() }
	}
	_ = vm.Status.Set("Ready")
	_ = vm.Name.Set(fullName)
	return vm
}

func (vm *VM) Load() {
	vm.mu.Lock()
	if vm.cancel != nil {
		vm.cancel()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	vm.cancel = cancel
	vm.mu.Unlock()

	vm.runOnMain(func() {
		_ = vm.Loading.Set(true)
		_ = vm.Error.Set("")
		_ = vm.Status.Set("Loading...")
	})

	go func() {
		details, err := vm.svc.LoadDetails(ctx, vm.FullName, vm.Token)
		if err != nil {
			vm.runOnMain(func() {
				_ = vm.Loading.Set(false)
				_ = vm.Error.Set(errMessage(err))
				_ = vm.Status.Set("Load failed")
			})
			return
		}

		vm.runOnMain(func() {
			vm.apply(details)
			_ = vm.Loading.Set(false)
			_ = vm.Status.Set("Loaded")
		})
	}()
}

func (vm *VM) apply(repo domain.RepoDetails) {
	_ = vm.Name.Set(valueOrDash(repo.FullName))
	_ = vm.Description.Set(valueOrDash(repo.Description))
	_ = vm.Language.Set(valueOrDash(repo.Language))
	_ = vm.Homepage.Set(valueOrDash(repo.Homepage))
	_ = vm.DefaultBranch.Set(valueOrDash(repo.DefaultBranch))
	_ = vm.License.Set(valueOrDash(repo.License))
	_ = vm.Topics.Set(valueOrDash(strings.Join(repo.Topics, ", ")))
	_ = vm.Stars.Set(fmt.Sprintf("%d", repo.Stars))
	_ = vm.Forks.Set(fmt.Sprintf("%d", repo.Forks))
	_ = vm.Watchers.Set(fmt.Sprintf("%d", repo.Watchers))
	_ = vm.OpenIssues.Set(fmt.Sprintf("%d", repo.OpenIssues))
	_ = vm.Size.Set(fmt.Sprintf("%d", repo.Size))
	_ = vm.UpdatedAt.Set(formatTime(repo.UpdatedAt))
	_ = vm.CreatedAt.Set(formatTime(repo.CreatedAt))
	_ = vm.PushedAt.Set(formatTime(repo.PushedAt))
	_ = vm.Private.Set(boolLabel(repo.Private))
	_ = vm.HTMLURL.Set(valueOrDash(repo.HTMLURL))
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("2006-01-02 15:04")
}

func valueOrDash(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "-"
	}
	return value
}

func boolLabel(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func errMessage(err error) string {
	if errors.Is(err, context.Canceled) {
		return "request canceled"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "request timeout"
	}
	return err.Error()
}
