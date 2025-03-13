package product

import (
	"context"

	"github.com/ahmetkoprulu/rtrp/common/data"
	"github.com/ahmetkoprulu/rtrp/models"
)

type ProductType string

const (
	ProductTypeChips ProductType = "chips"
)

type ProductStore interface {
	GetProduct(ctx context.Context, id string) (*models.Product, error)
	GiveRewardToPlayer(ctx context.Context, items []models.Item, playerID string) error
	GiveRewardToPlayerWithTx(ctx context.Context, tx data.QueryRunner, items []models.Item, playerID string) error
}
