package stars

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2/data/binding"

	"github.com/TBXark/gh-stars/internal/app/stars"
	"github.com/TBXark/gh-stars/internal/domain"
)

type VM struct {
	Username binding.String
	Token    binding.String
	PerPage  binding.String

	Loading binding.Bool
	Status  binding.String
	Error   binding.String

	Repos binding.UntypedList

	svc       stars.Loader
	runOnMain func(func())

	mu     sync.Mutex
	cancel context.CancelFunc

	reposMu sync.RWMutex
	repos   []domain.Repo
}

func NewVM(svc stars.Loader, runOnMain func(func())) *VM {
	vm := &VM{
		Username:  binding.NewString(),
		Token:     binding.NewString(),
		PerPage:   binding.NewString(),
		Loading:   binding.NewBool(),
		Status:    binding.NewString(),
		Error:     binding.NewString(),
		Repos:     binding.NewUntypedList(),
		svc:       svc,
		runOnMain: runOnMain,
	}
	if vm.runOnMain == nil {
		vm.runOnMain = func(f func()) { f() }
	}
	_ = vm.PerPage.Set("100")
	_ = vm.Status.Set("Ready")
	return vm
}

func (vm *VM) Load() {
	vm.mu.Lock()
	if vm.cancel != nil {
		vm.cancel()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	vm.cancel = cancel
	vm.mu.Unlock()

	vm.runOnMain(func() {
		_ = vm.Loading.Set(true)
		_ = vm.Error.Set("")
		_ = vm.Status.Set("Loading...")
	})

	go func() {
		username, _ := vm.Username.Get()
		token, _ := vm.Token.Get()
		perPageStr, _ := vm.PerPage.Get()
		perPage, err := parsePerPage(perPageStr)
		if err != nil {
			vm.runOnMain(func() {
				_ = vm.Loading.Set(false)
				_ = vm.Error.Set(err.Error())
				_ = vm.Status.Set("Load failed")
			})
			return
		}

		repos, err := vm.svc.LoadStarred(ctx, username, token, perPage)
		if err != nil {
			vm.runOnMain(func() {
				_ = vm.Loading.Set(false)
				_ = vm.Error.Set(err.Error())
				_ = vm.Status.Set("Load failed")
			})
			return
		}

		vm.runOnMain(func() {
			vm.setRepos(repos)
			_ = vm.Loading.Set(false)
			_ = vm.Status.Set("Loaded")
		})
	}()
}

func (vm *VM) Cleanup() {
	vm.mu.Lock()
	if vm.cancel != nil {
		vm.cancel()
		vm.cancel = nil
	}
	vm.mu.Unlock()
}

func (vm *VM) Clear() {
	vm.Cleanup()

	vm.runOnMain(func() {
		vm.setRepos(nil)
		_ = vm.Error.Set("")
		_ = vm.Status.Set("Cleared")
	})
}

func (vm *VM) RepoAt(index int) (domain.Repo, bool) {
	vm.reposMu.RLock()
	defer vm.reposMu.RUnlock()
	if index < 0 || index >= len(vm.repos) {
		return domain.Repo{}, false
	}
	return vm.repos[index], true
}

func (vm *VM) setRepos(repos []domain.Repo) {
	vm.reposMu.Lock()
	vm.repos = make([]domain.Repo, len(repos))
	copy(vm.repos, repos)
	vm.reposMu.Unlock()

	_ = vm.Repos.Set(nil)
	for _, repo := range repos {
		_ = vm.Repos.Append(repo)
	}
}

func parsePerPage(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 100, nil
	}
	perPage, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.New("per page must be a number")
	}
	if perPage < 1 || perPage > 100 {
		return 0, errors.New("per page must be between 1 and 100")
	}
	return perPage, nil
}
