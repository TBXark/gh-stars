package stars

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/TBXark/gh-stars/internal/domain"
)

func NewRepoList(vm *VM, onOpen func(domain.Repo)) fyne.CanvasObject {
	headers := container.NewGridWithColumns(5,
		headerLabel("Name", fyne.TextAlignLeading),
		headerLabel("Description", fyne.TextAlignLeading),
		headerLabel("Language", fyne.TextAlignLeading),
		headerLabel("Stars", fyne.TextAlignTrailing),
		headerLabel("Updated", fyne.TextAlignTrailing),
	)

	list := widget.NewListWithData(vm.Repos, func() fyne.CanvasObject {
		return newRepoRow()
	}, func(di binding.DataItem, obj fyne.CanvasObject) {
		repo, err := repoFromItem(di)
		if err != nil {
			return
		}
		updateRepoRow(obj, repo)
	})

	list.OnSelected = func(id widget.ListItemID) {
		repo, ok := vm.RepoAt(id)
		if ok && onOpen != nil {
			onOpen(repo)
		}
		list.Unselect(id)
	}

	header := container.NewVBox(headers, widget.NewSeparator())
	return container.NewBorder(header, nil, nil, nil, list)
}

func headerLabel(text string, align fyne.TextAlign) *widget.Label {
	return widget.NewLabelWithStyle(text, align, fyne.TextStyle{Bold: true})
}

func newRepoRow() fyne.CanvasObject {
	name := widget.NewLabel("")
	name.Wrapping = fyne.TextTruncate

	desc := widget.NewLabel("")
	desc.Wrapping = fyne.TextTruncate

	lang := widget.NewLabel("")
	lang.Wrapping = fyne.TextTruncate

	stars := widget.NewLabel("")
	stars.Alignment = fyne.TextAlignTrailing

	updated := widget.NewLabel("")
	updated.Alignment = fyne.TextAlignTrailing

	row := container.NewGridWithColumns(5, name, desc, lang, stars, updated)
	row.Layout = layout.NewGridLayoutWithColumns(5)
	return row
}

func updateRepoRow(obj fyne.CanvasObject, repo domain.Repo) {
	row, ok := obj.(*fyne.Container)
	if !ok || len(row.Objects) < 5 {
		return
	}
	setLabel(row.Objects[0], repo.FullName)
	setLabel(row.Objects[1], valueOrDash(repo.Description))
	setLabel(row.Objects[2], valueOrDash(repo.Language))
	setLabel(row.Objects[3], fmt.Sprintf("%d", repo.Stars))
	setLabel(row.Objects[4], formatDate(repo.UpdatedAt))
}

func setLabel(obj fyne.CanvasObject, text string) {
	label, ok := obj.(*widget.Label)
	if !ok {
		return
	}
	label.SetText(text)
}

func repoFromItem(di binding.DataItem) (domain.Repo, error) {
	item, ok := di.(binding.Item[domain.Repo])
	if !ok {
		return domain.Repo{}, fmt.Errorf("invalid list item")
	}
	return item.Get()
}

func valueOrDash(value string) string {
	if value == "" {
		return "-"
	}
	return value
}

func formatDate(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("2006-01-02")
}
