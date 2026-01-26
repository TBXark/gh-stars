package main

import (
	"fyne.io/fyne/v2/app"

	"github.com/tbxark/gh-stars/internal/app/repos"
	"github.com/tbxark/gh-stars/internal/app/stars"
	"github.com/tbxark/gh-stars/internal/github"
	"github.com/tbxark/gh-stars/internal/ui/nav"
)

func main() {
	fyneApp := app.New()

	client := github.NewClient(nil)
	starsSvc := stars.Service{GH: client}
	repoSvc := repos.Service{GH: client}

	router := &nav.AppNavigator{App: fyneApp, RepoSvc: repoSvc, StarsSvc: starsSvc}
	router.ShowStars()
	fyneApp.Run()
}
