package orders

import (
	repo "ecom/internal/adapters/postgresql/sqlc"
	"context"
)

type Service interface {
	CreateOrder(ctx context.Context, arg repo.CreateOrderParams) (repo.Order, error)
}

type svc struct {
	repo repo.Queries
}

func (s *svc) CreateOrder(ctx context.Context, arg repo.CreateOrderParams) (repo.Order, error) {
	return s.repo.CreateOrder(ctx, arg)
}
