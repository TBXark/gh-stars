package domain

import "time"

type Repo struct {
	FullName    string
	HTMLURL     string
	Description string
	Language    string
	Stars       int
	Forks       int
	UpdatedAt   time.Time
	Private     bool
}

type RepoDetails struct {
	FullName      string
	HTMLURL       string
	Description   string
	Language      string
	Homepage      string
	DefaultBranch string
	License       string
	Topics        []string
	Stars         int
	Forks         int
	Watchers      int
	OpenIssues    int
	Size          int
	UpdatedAt     time.Time
	CreatedAt     time.Time
	PushedAt      time.Time
	Private       bool
}
