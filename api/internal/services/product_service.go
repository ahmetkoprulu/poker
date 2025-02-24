package services

import (
	"context"

	"github.com/ahmetkoprulu/rtrp/common/data"
	"github.com/ahmetkoprulu/rtrp/internal/services/product"
	"github.com/ahmetkoprulu/rtrp/models"
)

type ProductService struct {
	productStore *product.PgProductStore
	db           *data.PgDbContext
}

func NewProductService(db *data.PgDbContext) *ProductService {
	return &ProductService{
		db:           db,
		productStore: product.NewPgProductStore(db),
	}
}

func (s *ProductService) GetProduct(ctx context.Context, id string) (*models.Product, error) {
	return s.productStore.GetProduct(ctx, id)
}

func (s *ProductService) GiveRewardToPlayer(ctx context.Context, items []models.Item, playerID string) error {
	return s.productStore.GiveRewardToPlayer(ctx, items, playerID)
}
