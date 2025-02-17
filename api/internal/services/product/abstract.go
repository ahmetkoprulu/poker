package product

import (
	"context"

	"github.com/ahmetkoprulu/rtrp/models"
)

type ProductType string

const (
	ProductTypeChips ProductType = "chips"
)

type ProductStore interface {
	GetProduct(ctx context.Context, id string) (*models.Product, error)
	GiveReward(ctx context.Context, playerID string, rewards []*models.Item) error
}
