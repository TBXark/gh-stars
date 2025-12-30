package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/TBXark/gh-stars/internal/app/repos"
	"github.com/TBXark/gh-stars/internal/app/stars"
	"github.com/TBXark/gh-stars/internal/github"
	"github.com/TBXark/gh-stars/internal/ui/nav"
	starsui "github.com/TBXark/gh-stars/internal/ui/stars"
)

func main() {
	fyneApp := app.New()
	w := fyneApp.NewWindow("GitHub Stars")
	w.Resize(fyne.NewSize(1100, 700))

	client := github.NewClient(nil)
	starsSvc := stars.Service{GH: client}
	repoSvc := repos.Service{GH: client}

	navigator := &nav.AppNavigator{App: fyneApp, RepoSvc: repoSvc}
	vm := starsui.NewVM(starsSvc, fyne.Do)
	w.SetContent(starsui.NewView(w, vm, navigator))

	w.ShowAndRun()
}
