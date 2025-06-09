package product

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ahmetkoprulu/rtrp/common/data"
	"github.com/ahmetkoprulu/rtrp/models"
)

// ItemHandler defines the interface for handling different types of items
type ItemHandler interface {
	HandleReward(ctx context.Context, tx data.QueryRunner, playerID string, item models.Item) error
}

type ChipsHandler struct{}

func (h *ChipsHandler) HandleReward(ctx context.Context, tx data.QueryRunner, playerID string, item models.Item) error {
	updateQuery := `
		UPDATE players
		SET chips = chips + $1
		WHERE id = $2
	`
	if _, err := tx.Exec(ctx, updateQuery, item.Amount, playerID); err != nil {
		return fmt.Errorf("failed to update player chips: %w", err)
	}
	return nil
}

type GoldHandler struct{}

func (h *GoldHandler) HandleReward(ctx context.Context, tx data.QueryRunner, playerID string, item models.Item) error {
	updateQuery := `
		UPDATE players
		SET golds = golds + $1
		WHERE id = $2
	`
	if _, err := tx.Exec(ctx, updateQuery, item.Amount, playerID); err != nil {
		return fmt.Errorf("failed to update player golds: %w", err)
	}
	return nil
}

type EventHandler struct{}

func (h *EventHandler) HandleReward(ctx context.Context, tx data.QueryRunner, playerID string, item models.Item) error {
	var metadata models.EventItemMetadata
	var metadataBytes, err = json.Marshal(item.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal event item metadata: %w", err)
	}

	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		return fmt.Errorf("failed to unmarshal event item metadata: %w", err)
	}
	eventID := metadata.EventID
	if eventID == "" {
		// return fmt.Errorf("event ID is empty")
		return nil
	}

	query := `
		UPDATE player_events
		SET tickets = tickets + $3
		WHERE player_id = $1 AND event_id = $2 AND expires_at > NOW()
	`

	if _, err := tx.Exec(ctx, query, eventID, playerID, item.Amount); err != nil {
		return fmt.Errorf("failed to update player event tickets: %w", err)
	}
	return nil
}

type SpinHandler struct{}

// TODO: SPin column does not exists.
func (h *SpinHandler) HandleReward(ctx context.Context, tx data.QueryRunner, playerID string, item models.Item) error {
	updateQuery := `
		UPDATE players
		SET spins = spins + $1
		WHERE id = $2
	`
	if _, err := tx.Exec(ctx, updateQuery, item.Amount, playerID); err != nil {
		return fmt.Errorf("failed to update player spins: %w", err)
	}
	return nil
}

type GoldSpinHandler struct{}

func (h *GoldSpinHandler) HandleReward(ctx context.Context, tx data.QueryRunner, playerID string, item models.Item) error {
	updateQuery := `
		UPDATE players
		SET gold_spins = gold_spins + $1
		WHERE id = $2
	`
	if _, err := tx.Exec(ctx, updateQuery, item.Amount, playerID); err != nil {
		return fmt.Errorf("failed to update player gold spins: %w", err)
	}
	return nil
}

type BattlePassXPHandler struct{}

func (h *BattlePassXPHandler) HandleReward(ctx context.Context, tx data.QueryRunner, playerID string, item models.Item) error {
	activeBattlePassQuery := `
		SELECT id
		FROM battle_passes
		WHERE start_time <= NOW() AND end_time >= NOW() AND status = 'active'
		LIMIT 1
	`

	var activeBattlePassID string
	err := tx.QueryRow(ctx, activeBattlePassQuery).Scan(&activeBattlePassID)
	if err != nil {
		return fmt.Errorf("failed to get active battle pass: %w", err)
	}

	if item.Amount <= 0 {
		// return fmt.Errorf("invalid XP amount: %d", item.Amount)
		return nil
	}

	query := `
		SELECT pbp.id, pbp.battle_pass_id, pbp.current_level, pbp.current_xp, bp.max_level
		FROM player_battle_passes pbp
		JOIN battle_passes bp ON pbp.battle_pass_id = bp.id
		WHERE pbp.battle_pass_id = $1 AND pbp.player_id = $2
		FOR UPDATE
	`

	var currentLevel, currentXP, maxLevel int
	var battlePassID, playerBattlePassID string
	err = tx.QueryRow(ctx, query, activeBattlePassID, playerID).Scan(&playerBattlePassID, &battlePassID, &currentLevel, &currentXP, &maxLevel)
	if err != nil {
		return err
	}

	newXP := currentXP + item.Amount
	newLevel := currentLevel

	for newLevel < maxLevel {
		var requiredXP int
		query = `
			SELECT required_xp
			FROM battle_pass_levels
			WHERE battle_pass_id = $1
			AND level = $2
		`

		err = tx.QueryRow(ctx, query, battlePassID, newLevel+1).Scan(&requiredXP)
		if err != nil {
			break // No more levels found
		}

		if newXP >= requiredXP {
			newLevel++
			newXP -= requiredXP
		} else {
			break
		}
	}

	query = `
		UPDATE player_battle_passes
		SET current_level = $1,
			current_xp = $2,
			updated_at = NOW()
		WHERE id = $3
	`

	_, err = tx.Exec(ctx, query, newLevel, newXP, playerBattlePassID)
	if err != nil {
		return err
	}

	return nil
}

type MultiplierHandler struct{}

func (h *MultiplierHandler) HandleReward(ctx context.Context, tx data.QueryRunner, playerID string, item models.Item) error {
	// var metadata models.MultiplierItemMetadata
	// var metadataBytes, err = json.Marshal(item.Metadata)
	// if err != nil {
	// 	return fmt.Errorf("failed to marshal multiplier item metadata: %w", err)
	// }

	// if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
	// 	return fmt.Errorf("failed to unmarshal multiplier item metadata: %w", err)
	// }
	// eventID := metadata.EventID

	// query := `
	// 	UPDATE player_events
	// 	SET multiplier = multiplier + $1
	// 	WHERE id = $2 AND event_id = $3
	// `

	// _, err = tx.Exec(ctx, query, item.Amount, eventID)
	// if err != nil {
	// 	return fmt.Errorf("failed to update player event multiplier: %w", err)
	// }

	return nil
}

type PgProductStore struct {
	db           *data.PgDbContext
	itemHandlers map[models.ItemType]ItemHandler
}

func NewPgProductStore(db *data.PgDbContext) *PgProductStore {
	store := &PgProductStore{
		db:           db,
		itemHandlers: make(map[models.ItemType]ItemHandler),
	}

	// Register handlers for different item types
	store.itemHandlers[models.ItemTypeChips] = &ChipsHandler{}
	store.itemHandlers[models.ItemTypeGold] = &GoldHandler{}
	store.itemHandlers[models.ItemTypeSpin] = &SpinHandler{}
	store.itemHandlers[models.ItemTypeGoldSpin] = &GoldSpinHandler{}
	store.itemHandlers[models.ItemTypeEvent] = &EventHandler{}
	store.itemHandlers[models.ItemTypeMultiplier] = &MultiplierHandler{}
	store.itemHandlers[models.ItemTypeBattlePassXP] = &BattlePassXPHandler{}

	return store
}

func (s *PgProductStore) GetProduct(ctx context.Context, id string) (*models.Product, error) {
	query := `
		SELECT id, item, price
		FROM products
		WHERE id = $1
	`

	var product models.Product
	err := s.db.QueryRow(ctx, query, id).Scan(&product.ID, &product.Item, &product.Price)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (s *PgProductStore) GiveRewardToPlayerWithTx(ctx context.Context, tx data.QueryRunner, items []models.Item, playerID string) error {
	for _, item := range items {
		if handler, exists := s.itemHandlers[item.Type]; exists {
			if err := handler.HandleReward(ctx, tx, playerID, item); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("no handler found for item type: %v", item.Type)
		}
	}

	return nil
}

func (s *PgProductStore) GiveRewardToPlayer(ctx context.Context, items []models.Item, playerID string) error {
	return s.db.WithTransaction(ctx, func(tx data.QueryRunner) error {
		return s.GiveRewardToPlayerWithTx(ctx, tx, items, playerID)
	})
}
