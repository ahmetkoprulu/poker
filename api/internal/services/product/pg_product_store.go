package product

// import (
// 	"context"
// 	"fmt"

// 	"github.com/ahmetkoprulu/rtrp/common/data"
// 	"github.com/ahmetkoprulu/rtrp/models"
// 	"github.com/jackc/pgx/v5"
// )

// // ItemHandler defines the interface for handling different types of items
// type ItemHandler interface {
// 	HandleReward(ctx context.Context, tx *pgx.Tx, playerID string, item *models.Item) error
// }

// // ChipsHandler handles chip rewards
// type ChipsHandler struct{}

// func (h *ChipsHandler) HandleReward(ctx context.Context, tx *pgx.Tx, playerID string, item *models.Item) error {
// 	updateQuery := `
// 		UPDATE players
// 		SET chips = chips + $1
// 		WHERE id = $2
// 	`
// 	if _, err := tx.Exec(ctx, updateQuery, item.Amount, playerID); err != nil {
// 		return fmt.Errorf("failed to update player chips: %w", err)
// 	}
// 	return nil
// }

// // GoldHandler handles gold rewards (example of another handler)
// type GoldHandler struct{}

// func (h *GoldHandler) HandleReward(ctx context.Context, tx *pgx.Tx, playerID string, item *models.Item) error {
// 	updateQuery := `
// 		UPDATE players
// 		SET gold = gold + $1
// 		WHERE id = $2
// 	`
// 	if _, err := tx.Exec(ctx, updateQuery, item.Amount, playerID); err != nil {
// 		return fmt.Errorf("failed to update player gold: %w", err)
// 	}
// 	return nil
// }

// type PgProductStore struct {
// 	db           *data.PgDbContext
// 	itemHandlers map[models.ItemType]ItemHandler
// }

// func NewPgProductStore(db *data.PgDbContext) *PgProductStore {
// 	store := &PgProductStore{
// 		db:           db,
// 		itemHandlers: make(map[models.ItemType]ItemHandler),
// 	}

// 	// Register handlers for different item types
// 	store.itemHandlers[models.ItemTypeChips] = &ChipsHandler{}
// 	store.itemHandlers[models.ItemTypeGold] = &GoldHandler{}

// 	return store
// }

// func (s *PgProductStore) GetProduct(ctx context.Context, id string) (*models.Product, error) {
// 	query := `
// 		SELECT id, item, price
// 		FROM products
// 		WHERE id = $1
// 	`

// 	var product models.Product
// 	err := s.db.QueryRow(ctx, query, id).Scan(&product.ID, &product.Item, &product.Price)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &product, nil
// }

// func (s *PgProductStore) GiveRewardToPlayer(ctx context.Context, items []*models.Item, playerID string) error {
// 	tx, err := s.db.Begin(ctx)
// 	if err != nil {
// 		return fmt.Errorf("failed to begin transaction: %w", err)
// 	}
// 	defer tx.Rollback(ctx)

// 	// Insert rewards into product_rewards table
// 	rewardQuery := `
// 		INSERT INTO product_rewards (player_id, product_id, amount)
// 		VALUES ($1, $2, $3)
// 	`

// 	for _, item := range items {
// 		// Record the reward
// 		if _, err := tx.Exec(ctx, rewardQuery, playerID, item.Metadata["product_id"], item.Amount); err != nil {
// 			return fmt.Errorf("failed to insert reward record: %w", err)
// 		}

// 		// Handle the reward using appropriate handler
// 		if handler, exists := s.itemHandlers[item.Type]; exists {
// 			if err := handler.HandleReward(ctx, tx, playerID, item); err != nil {
// 				return err
// 			}
// 		} else {
// 			return fmt.Errorf("no handler found for item type: %v", item.Type)
// 		}
// 	}

// 	if err := tx.Commit(ctx); err != nil {
// 		return fmt.Errorf("failed to commit transaction: %w", err)
// 	}
// 	return nil
// }
