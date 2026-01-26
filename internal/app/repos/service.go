package repos

import (
	"context"

	"github.com/tbxark/gh-stars/internal/domain"
	"github.com/tbxark/gh-stars/internal/github"
)

// Loader defines the interface for loading repository details
type Loader interface {
	LoadDetails(ctx context.Context, fullName, token string) (domain.RepoDetails, error)
}

type Service struct {
	GH github.Client
}

func (s Service) LoadDetails(ctx context.Context, fullName, token string) (domain.RepoDetails, error) {
	return s.GH.GetRepoDetails(ctx, fullName, token)
}
