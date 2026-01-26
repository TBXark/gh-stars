package stars

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/tbxark/gh-stars/internal/domain"
	"github.com/tbxark/gh-stars/internal/ui/route"
	"github.com/tbxark/gh-stars/internal/ui/widgets"
)

func NewView(w fyne.Window, vm *VM, router route.Router) fyne.CanvasObject {
	loadBtn := widget.NewButtonWithIcon("Load Stars", theme.DownloadIcon(), vm.Load)
	clearBtn := widget.NewButtonWithIcon("Clear", theme.ContentClearIcon(), vm.Clear)

	vm.Loading.AddListener(binding.NewDataListener(func() {
		loading, _ := vm.Loading.Get()
		if loading {
			loadBtn.Disable()
		} else {
			loadBtn.Enable()
		}
	}))

	form := widgets.NewCredentialsForm(vm.Username, vm.Token, vm.PerPage)
	credentialsCard := widget.NewCard(
		"",
		"Token is optional but improves rate limits.",
		form,
	)
	statusBar := widgets.NewStatusPanel(vm.Status, vm.Error, vm.Loading)

	title := canvas.NewText("GitHub Stars", theme.ForegroundColor())
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = theme.TextHeadingSize()

	subtitle := canvas.NewText("Browse and open your starred repositories.", theme.DisabledColor())
	subtitle.TextSize = theme.TextSubHeadingSize()

	actionBar := container.NewHBox(layout.NewSpacer(), loadBtn, clearBtn)
	header := container.NewBorder(nil, nil, nil, actionBar, container.NewVBox(title, subtitle))

	onOpen := func(repo domain.Repo) {
		if router == nil {
			return
		}
		tokenStr, _ := vm.Token.Get()
		router.ShowRepoDetails(repo.FullName, tokenStr)
	}

	list := NewRepoList(vm, onOpen)
	listCard := widget.NewCard(
		"",
		"Select a repo to open details.",
		list,
	)

	top := container.NewVBox(header, widget.NewSeparator(), credentialsCard)
	return container.NewBorder(
		container.NewPadded(top),
		container.NewPadded(statusBar),
		nil,
		nil,
		container.NewPadded(listCard),
	)
}
