package stars

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
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
		return newRepoRowWidget()
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

func updateRepoRow(obj fyne.CanvasObject, repo domain.Repo) {
	row, ok := obj.(*repoRowWidget)
	if !ok {
		return
	}
	row.name.SetText(repo.FullName)
	row.desc.SetText(valueOrDash(repo.Description))
	row.lang.SetText(valueOrDash(repo.Language))
	row.stars.SetText(fmt.Sprintf("%d", repo.Stars))
	row.updated.SetText(formatDate(repo.UpdatedAt))
}

type repoRowWidget struct {
	widget.BaseWidget
	name    *widget.Label
	desc    *widget.Label
	lang    *widget.Label
	stars   *widget.Label
	updated *widget.Label
}

func newRepoRowWidget() *repoRowWidget {
	row := &repoRowWidget{
		name:    widget.NewLabel(""),
		desc:    widget.NewLabel(""),
		lang:    widget.NewLabel(""),
		stars:   widget.NewLabel(""),
		updated: widget.NewLabel(""),
	}
	row.name.Wrapping = fyne.TextTruncate
	row.desc.Wrapping = fyne.TextTruncate
	row.lang.Wrapping = fyne.TextTruncate
	row.stars.Alignment = fyne.TextAlignTrailing
	row.updated.Alignment = fyne.TextAlignTrailing
	row.ExtendBaseWidget(row)
	return row
}

func (row *repoRowWidget) CreateRenderer() fyne.WidgetRenderer {
	grid := container.NewGridWithColumns(5, row.name, row.desc, row.lang, row.stars, row.updated)
	return widget.NewSimpleRenderer(grid)
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
