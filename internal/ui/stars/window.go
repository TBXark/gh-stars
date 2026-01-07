package stars

import (
	"fyne.io/fyne/v2"

	appstars "github.com/TBXark/gh-stars/internal/app/stars"
	"github.com/TBXark/gh-stars/internal/ui/route"
)

func NewStarsWindow(app fyne.App, svc appstars.Loader, router route.Router) fyne.Window {
	w := app.NewWindow("GitHub Stars")
	w.Resize(fyne.NewSize(1100, 700))

	vm := NewVM(svc, fyne.Do)
	w.SetContent(NewView(w, vm, router))
	w.SetOnClosed(func() {
		vm.Cleanup()
	})

	return w
}
