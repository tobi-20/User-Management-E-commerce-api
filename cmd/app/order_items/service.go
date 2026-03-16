package order_items

import (
	repo "ecom/internal/adapters/postgresql/sqlc"
	"context"
)

type Service interface {
	CreateOrderItems(ctx context.Context, arg repo.CreateOrderItemParams) (repo.OrderItem, error)
}

type svc struct {
	repo repo.Queries
}

func (s *svc) CreateOrderItems(ctx context.Context, arg repo.CreateOrderItemParams) (repo.OrderItem, error) {
	return s.repo.CreateOrderItem(ctx, arg)
}
