package brand

import (
	repo "ecom/internal/adapters/postgresql/sqlc"
	"context"
)

type Service interface {
	CreateBrand(ctx context.Context, name string) (repo.Brand, error)
}

type svc struct {
	repo repo.Querier
}

func NewService(repo repo.Querier) Service {
	return &svc{
		repo: repo,
	}
}

func (s *svc) CreateBrand(ctx context.Context, name string) (repo.Brand, error) {
	return s.repo.CreateBrand(ctx, name)
}
