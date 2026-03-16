package categories

import (
	repo "ecom/internal/adapters/postgresql/sqlc"
	"context"
)

type Service interface {
	CreateCategory(ctx context.Context, name string) (repo.Category, error)
}

type svc struct {
	repo repo.Queries
}

func NewService(repo repo.Queries) Service {
	return &svc{
		repo: repo,
	}
}

func (s *svc) CreateCategory(ctx context.Context, name string) (repo.Category, error) {
	return s.repo.CreateCategory(ctx, name)
}
