package details

import (
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/TBXark/gh-stars/internal/ui/widgets"
)

func NewView(w fyne.Window, vm *VM) fyne.CanvasObject {
	refresh := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), vm.Load)

	vm.Loading.AddListener(binding.NewDataListener(func() {
		loading, _ := vm.Loading.Get()
		if loading {
			refresh.Disable()
		} else {
			refresh.Enable()
		}
	}))

	openBtn := widget.NewButtonWithIcon("Open in Browser", theme.NavigateNextIcon(), func() {
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

	title := canvas.NewText("Repository Details", theme.ForegroundColor())
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = theme.TextHeadingSize()

	name := widget.NewLabelWithData(vm.Name)
	name.Wrapping = fyne.TextTruncate

	newWrapLabel := func(value binding.String) *widget.Label {
		label := widget.NewLabelWithData(value)
		label.Wrapping = fyne.TextWrapBreak
		return label
	}

	form := widget.NewForm(
		widget.NewFormItem("Full Name", newWrapLabel(vm.Name)),
		widget.NewFormItem("Description", newWrapLabel(vm.Description)),
		widget.NewFormItem("Language", newWrapLabel(vm.Language)),
		widget.NewFormItem("Stars", newWrapLabel(vm.Stars)),
		widget.NewFormItem("Forks", newWrapLabel(vm.Forks)),
		widget.NewFormItem("Watchers", newWrapLabel(vm.Watchers)),
		widget.NewFormItem("Open Issues", newWrapLabel(vm.OpenIssues)),
		widget.NewFormItem("Default Branch", newWrapLabel(vm.DefaultBranch)),
		widget.NewFormItem("License", newWrapLabel(vm.License)),
		widget.NewFormItem("Topics", newWrapLabel(vm.Topics)),
		widget.NewFormItem("Homepage", newWrapLabel(vm.Homepage)),
		widget.NewFormItem("HTML URL", newWrapLabel(vm.HTMLURL)),
		widget.NewFormItem("Private", newWrapLabel(vm.Private)),
		widget.NewFormItem("Size", newWrapLabel(vm.Size)),
		widget.NewFormItem("Updated", newWrapLabel(vm.UpdatedAt)),
		widget.NewFormItem("Created", newWrapLabel(vm.CreatedAt)),
		widget.NewFormItem("Pushed", newWrapLabel(vm.PushedAt)),
	)

	header := container.NewBorder(
		nil,
		nil,
		nil,
		container.NewHBox(layout.NewSpacer(), openBtn, refresh),
		container.NewVBox(title, name),
	)
	statusBar := widgets.NewStatusPanel(vm.Status, vm.Error, vm.Loading)

	detailsCard := widget.NewCard("", "Overview and metadata.", form)
	content := container.NewVScroll(container.NewPadded(detailsCard))
	top := container.NewPadded(container.NewVBox(header, widget.NewSeparator()))

	return container.NewBorder(top, container.NewPadded(statusBar), nil, nil, content)
}
