package product_variants

import (
	repo "ecom/internal/adapters/postgresql/sqlc"
	"context"
)

type Service interface {
	CreateProductVariant(ctx context.Context, arg repo.CreateProductVariantParams) (repo.ProductVariant, error)
}

type svc struct {
	repo repo.Queries
}

func (s *svc) CreateProductVariant(ctx context.Context, arg repo.CreateProductVariantParams) (repo.ProductVariant, error) {
	return s.repo.CreateProductVariant(ctx, arg)
}
