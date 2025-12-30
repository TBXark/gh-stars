package nav

import (
	"sync"

	"fyne.io/fyne/v2"

	"github.com/TBXark/gh-stars/internal/app/repos"
	"github.com/TBXark/gh-stars/internal/ui/details"
)

type Navigator interface {
	ShowRepoDetails(fullName, token string)
}

type AppNavigator struct {
	App     fyne.App
	RepoSvc repos.Service

	mu      sync.Mutex
	details map[string]fyne.Window
}

func (n *AppNavigator) ShowRepoDetails(fullName, token string) {
	n.mu.Lock()
	if n.details == nil {
		n.details = map[string]fyne.Window{}
	}
	if w, ok := n.details[fullName]; ok {
		n.mu.Unlock()
		w.RequestFocus()
		w.Show()
		return
	}
	n.mu.Unlock()

	w := details.NewRepoDetailsWindow(n.App, n.RepoSvc, fullName, token)

	n.mu.Lock()
	n.details[fullName] = w
	n.mu.Unlock()

	w.SetOnClosed(func() {
		n.mu.Lock()
		delete(n.details, fullName)
		n.mu.Unlock()
	})
	w.Show()
}
