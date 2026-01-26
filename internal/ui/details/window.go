package details

import (
	"fyne.io/fyne/v2"

	"github.com/tbxark/gh-stars/internal/app/repos"
)

func NewRepoDetailsWindow(app fyne.App, svc repos.Loader, fullName, token string) fyne.Window {
	w := app.NewWindow("Repo Details: " + fullName)
	w.Resize(fyne.NewSize(900, 600))

	vm := NewVM(svc, fullName, token, fyne.Do)
	w.SetContent(NewView(w, vm))
	w.SetOnClosed(func() {
		vm.Cleanup()
	})
	vm.Load()

	return w
}
