package details

import (
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/TBXark/gh-stars/internal/ui/widgets"
)

func NewView(w fyne.Window, vm *VM) fyne.CanvasObject {
	refresh := widget.NewButton("Refresh", vm.Load)

	vm.Loading.AddListener(binding.NewDataListener(func() {
		loading, _ := vm.Loading.Get()
		if loading {
			refresh.Disable()
		} else {
			refresh.Enable()
		}
	}))

	openBtn := widget.NewButton("Open in Browser", func() {
		urlStr, _ := vm.HTMLURL.Get()
		if urlStr == "" || urlStr == "-" {
			return
		}
		parsed, err := url.Parse(urlStr)
		if err != nil {
			return
		}
		_ = fyne.CurrentApp().OpenURL(parsed)
	})

	title := widget.NewLabelWithData(vm.Name)
	title.Wrapping = fyne.TextTruncate

	desc := widget.NewLabelWithData(vm.Description)
	desc.Wrapping = fyne.TextWrapWord

	form := widget.NewForm(
		widget.NewFormItem("Full Name", widget.NewLabelWithData(vm.Name)),
		widget.NewFormItem("Description", desc),
		widget.NewFormItem("Language", widget.NewLabelWithData(vm.Language)),
		widget.NewFormItem("Stars", widget.NewLabelWithData(vm.Stars)),
		widget.NewFormItem("Forks", widget.NewLabelWithData(vm.Forks)),
		widget.NewFormItem("Watchers", widget.NewLabelWithData(vm.Watchers)),
		widget.NewFormItem("Open Issues", widget.NewLabelWithData(vm.OpenIssues)),
		widget.NewFormItem("Default Branch", widget.NewLabelWithData(vm.DefaultBranch)),
		widget.NewFormItem("License", widget.NewLabelWithData(vm.License)),
		widget.NewFormItem("Topics", widget.NewLabelWithData(vm.Topics)),
		widget.NewFormItem("Homepage", widget.NewLabelWithData(vm.Homepage)),
		widget.NewFormItem("HTML URL", widget.NewLabelWithData(vm.HTMLURL)),
		widget.NewFormItem("Private", widget.NewLabelWithData(vm.Private)),
		widget.NewFormItem("Size", widget.NewLabelWithData(vm.Size)),
		widget.NewFormItem("Updated", widget.NewLabelWithData(vm.UpdatedAt)),
		widget.NewFormItem("Created", widget.NewLabelWithData(vm.CreatedAt)),
		widget.NewFormItem("Pushed", widget.NewLabelWithData(vm.PushedAt)),
	)

	top := container.NewHBox(title, layout.NewSpacer(), openBtn, refresh)
	statusBar := widgets.NewStatusPanel(vm.Status, vm.Error)

	return container.NewBorder(top, statusBar, nil, nil, form)
}
