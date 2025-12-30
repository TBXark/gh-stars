package repos

import (
	"context"

	"github.com/TBXark/gh-stars/internal/domain"
	"github.com/TBXark/gh-stars/internal/github"
)

type Service struct {
	GH github.Client
}

func (s Service) LoadDetails(ctx context.Context, fullName, token string) (domain.RepoDetails, error) {
	return s.GH.GetRepoDetails(ctx, fullName, token)
}
