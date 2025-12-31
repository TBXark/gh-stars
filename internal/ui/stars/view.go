package stars

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/TBXark/gh-stars/internal/domain"
	"github.com/TBXark/gh-stars/internal/ui/route"
)

func NewView(w fyne.Window, vm *VM, router route.Router) fyne.CanvasObject {
	username := widget.NewEntryWithData(vm.Username)
	username.SetPlaceHolder("octocat")

	token := widget.NewPasswordEntry()
	token.Bind(vm.Token)
	token.SetPlaceHolder("optional")

	perPage := widget.NewEntryWithData(vm.PerPage)
	perPage.SetPlaceHolder("1-100")

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

	form := widget.NewForm(
		widget.NewFormItem("Username", username),
		widget.NewFormItem("Token", token),
		widget.NewFormItem("per_page", perPage),
	)

	toolbar := container.NewHBox(loadBtn, clearBtn, layout.NewSpacer())
	status := widget.NewLabelWithData(vm.Status)
	errLabel := widget.NewLabelWithData(vm.Error)
	statusBar := container.NewVBox(status, errLabel)

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
