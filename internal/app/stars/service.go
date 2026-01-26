package stars

import (
	"context"

	"github.com/tbxark/gh-stars/internal/domain"
	"github.com/tbxark/gh-stars/internal/github"
)

// Loader defines the interface for loading starred repositories
type Loader interface {
	LoadStarred(ctx context.Context, username, token string, perPage int) ([]domain.Repo, error)
}

type Service struct {
	GH github.Client
}

func (s Service) LoadStarred(ctx context.Context, username, token string, perPage int) ([]domain.Repo, error) {
	return s.GH.ListStarred(ctx, username, token, perPage)
}
