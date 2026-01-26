package testdata

import (
	"time"

	"github.com/tbxark/gh-stars/internal/domain"
)

// SampleRepo returns a sample Repo for testing
func SampleRepo() domain.Repo {
	return domain.Repo{
		FullName:    "golang/go",
		HTMLURL:     "https://github.com/golang/go",
		Description: "The Go programming language",
		Language:    "Go",
		Stars:       123456,
		Forks:       12345,
		UpdatedAt:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Private:     false,
	}
}

// SampleRepoPrivate returns a private repo for testing
func SampleRepoPrivate() domain.Repo {
	return domain.Repo{
		FullName:    "user/private-repo",
		HTMLURL:     "https://github.com/user/private-repo",
		Description: "Private repository",
		Language:    "Go",
		Stars:       42,
		Forks:       5,
		UpdatedAt:   time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC),
		Private:     true,
	}
}

// SampleRepoList returns a list of repos for testing
func SampleRepoList() []domain.Repo {
	return []domain.Repo{
		SampleRepo(),
		{
			FullName:    "kubernetes/kubernetes",
			HTMLURL:     "https://github.com/kubernetes/kubernetes",
			Description: "Production-Grade Container Scheduling and Management",
			Language:    "Go",
			Stars:       98765,
			Forks:       32100,
			UpdatedAt:   time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
			Private:     false,
		},
		{
			FullName:    "microsoft/vscode",
			HTMLURL:     "https://github.com/microsoft/vscode",
			Description: "Visual Studio Code",
			Language:    "TypeScript",
			Stars:       154000,
			Forks:       27500,
			UpdatedAt:   time.Date(2024, 1, 20, 12, 0, 0, 0, time.UTC),
			Private:     false,
		},
	}
}

// SampleRepoDetails returns detailed repo info for testing
func SampleRepoDetails() domain.RepoDetails {
	return domain.RepoDetails{
		FullName:      "golang/go",
		HTMLURL:       "https://github.com/golang/go",
		Description:   "The Go programming language",
		Language:      "Go",
		Homepage:      "https://go.dev",
		DefaultBranch: "main",
		License:       "BSD 3-Clause",
		Topics:        []string{"go", "golang", "programming-language", "compiler"},
		Stars:         123456,
		Forks:         12345,
		Watchers:      5000,
		OpenIssues:    3500,
		Size:          250000,
		UpdatedAt:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		CreatedAt:     time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC),
		PushedAt:      time.Date(2024, 1, 1, 11, 30, 0, 0, time.UTC),
		Private:       false,
	}
}

// SampleRepoDetailsMinimal returns minimal repo details for testing edge cases
func SampleRepoDetailsMinimal() domain.RepoDetails {
	return domain.RepoDetails{
		FullName:      "user/minimal-repo",
		HTMLURL:       "https://github.com/user/minimal-repo",
		Description:   "",
		Language:      "",
		Homepage:      "",
		DefaultBranch: "master",
		License:       "",
		Topics:        []string{},
		Stars:         0,
		Forks:         0,
		Watchers:      0,
		OpenIssues:    0,
		Size:          10,
		UpdatedAt:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		CreatedAt:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		PushedAt:      time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Private:       false,
	}
}

// SampleRepoDetailsWithNilTopics returns details with nil topics for testing
func SampleRepoDetailsWithNilTopics() domain.RepoDetails {
	details := SampleRepoDetails()
	details.Topics = nil
	return details
}

// InvalidRepoFullName represents invalid repo names for error testing
var InvalidRepoFullName = []string{
	"",               // empty
	"no-slash",       // missing owner
	"/no-owner",      // missing owner
	"no-repo/",       // missing repo
	"too/many/parts", // too many slashes
}
