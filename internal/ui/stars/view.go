package stars

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/TBXark/gh-stars/internal/domain"
	"github.com/TBXark/gh-stars/internal/ui/route"
	"github.com/TBXark/gh-stars/internal/ui/widgets"
)

func NewView(w fyne.Window, vm *VM, router route.Router) fyne.CanvasObject {
	loadBtn := widget.NewButton("Load Stars", vm.Load)
	clearBtn := widget.NewButton("Clear", vm.Clear)

	vm.Loading.AddListener(binding.NewDataListener(func() {
		loading, _ := vm.Loading.Get()
		if loading {
			loadBtn.Disable()
		} else {
			loadBtn.Enable()
		}
	}))

	form := widgets.NewCredentialsForm(vm.Username, vm.Token, vm.PerPage)
	toolbar := container.NewHBox(loadBtn, clearBtn, layout.NewSpacer())
	statusBar := widgets.NewStatusPanel(vm.Status, vm.Error)

	onOpen := func(repo domain.Repo) {
		if router == nil {
			return
		}
		tokenStr, _ := vm.Token.Get()
		router.ShowRepoDetails(repo.FullName, tokenStr)
	}

	list := NewRepoList(vm, onOpen)
	top := container.NewVBox(form, toolbar)
	return container.NewBorder(top, statusBar, nil, nil, list)
}
