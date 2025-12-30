package stars

import (
	"context"

	"github.com/TBXark/gh-stars/internal/domain"
	"github.com/TBXark/gh-stars/internal/github"
)

type Service struct {
	GH github.Client
}

func (s Service) LoadStarred(ctx context.Context, username, token string, perPage int) ([]domain.Repo, error) {
	return s.GH.ListStarred(ctx, username, token, perPage)
}
