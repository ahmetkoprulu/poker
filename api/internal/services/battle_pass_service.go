package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/ahmetkoprulu/rtrp/common/data"
	"github.com/ahmetkoprulu/rtrp/models"
	"github.com/google/uuid"
)

var (
	ErrBattlePassNotFound   = errors.New("battle pass not found")
	ErrLevelNotFound        = errors.New("battle pass level not found")
	ErrInvalidXPAmount      = errors.New("invalid XP amount")
	ErrRewardAlreadyClaimed = errors.New("reward already claimed")
	ErrInsufficientLevel    = errors.New("insufficient level to claim reward")
	ErrPremiumRequired      = errors.New("premium battle pass required to claim this reward")
	ErrPlayerIDRequired     = errors.New("player_id_required")
)

type BattlePassService struct {
	db             *data.PgDbContext
	productService *ProductService // For handling rewards
}

func NewBattlePassService(db *data.PgDbContext, productService *ProductService) *BattlePassService {
	return &BattlePassService{
		db:             db,
		productService: productService,
	}
}

// CreateBattlePass creates a new battle pass season
func (s *BattlePassService) CreateBattlePass(ctx context.Context, battlePass *models.BattlePass) error {
	battlePass.ID = uuid.New().String()
	battlePass.CreatedAt = time.Now()
	battlePass.UpdatedAt = time.Now()

	query := `
		INSERT INTO battle_passes (
			id, name, description, start_time, end_time,
			status, max_level, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9
		)
	`

	_, err := s.db.Exec(ctx, query,
		battlePass.ID, battlePass.Name, battlePass.Description,
		battlePass.StartTime, battlePass.EndTime,
		battlePass.Status, battlePass.MaxLevel,
		battlePass.CreatedAt, battlePass.UpdatedAt,
	)

	return err
}

// CreateBattlePassLevel creates a new level for a battle pass
func (s *BattlePassService) CreateBattlePassLevel(ctx context.Context, level *models.BattlePassLevel) error {
	level.ID = uuid.New().String()
	level.CreatedAt = time.Now()

	query := `
		INSERT INTO battle_pass_levels (
			id, battle_pass_id, level, required_xp,
			free_rewards, premium_rewards, created_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7
		)
	`

	_, err := s.db.Exec(ctx, query,
		level.ID, level.BattlePassID, level.Level,
		level.RequiredXP, level.FreeRewards, level.PremiumRewards,
		level.CreatedAt,
	)

	return err
}

// GetActiveBattlePass gets the currently active battle pass
func (s *BattlePassService) GetActiveBattlePass(ctx context.Context) (*models.BattlePass, error) {
	query := `
		SELECT id, name, description, start_time, end_time,
			   status, max_level, created_at, updated_at
		FROM battle_passes
		WHERE status = $1
		AND start_time <= NOW()
		AND end_time >= NOW()
		LIMIT 1
	`

	battlePass := &models.BattlePass{}
	err := s.db.QueryRow(ctx, query, models.BattlePassStatusActive).Scan(
		&battlePass.ID, &battlePass.Name, &battlePass.Description,
		&battlePass.StartTime, &battlePass.EndTime,
		&battlePass.Status, &battlePass.MaxLevel,
		&battlePass.CreatedAt, &battlePass.UpdatedAt,
	)

	if err != nil {
		return nil, ErrBattlePassNotFound
	}

	return battlePass, nil
}

// GetOrCreatePlayerBattlePass gets or creates a player's battle pass progress
func (s *BattlePassService) GetOrCreatePlayerBattlePass(ctx context.Context, playerID string, battlePassID string) (*models.PlayerBattlePass, error) {
	// First try to get existing progress
	query := `
		SELECT id, player_id, battle_pass_id, current_level,
			   current_xp, is_premium, created_at, updated_at
		FROM player_battle_passes
		WHERE player_id = $1 AND battle_pass_id = $2
	`

	playerBattlePass := &models.PlayerBattlePass{}
	err := s.db.QueryRow(ctx, query, playerID, battlePassID).Scan(
		&playerBattlePass.ID, &playerBattlePass.PlayerID,
		&playerBattlePass.BattlePassID, &playerBattlePass.CurrentLevel,
		&playerBattlePass.CurrentXP, &playerBattlePass.IsPremium,
		&playerBattlePass.CreatedAt, &playerBattlePass.UpdatedAt,
	)

	if err == nil {
		return playerBattlePass, nil
	}

	// Create new progress if not found
	playerBattlePass = &models.PlayerBattlePass{
		ID:           uuid.New().String(),
		PlayerID:     playerID,
		BattlePassID: battlePassID,
		CurrentLevel: 1,
		CurrentXP:    0,
		IsPremium:    false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	query = `
		INSERT INTO player_battle_passes (
			id, player_id, battle_pass_id, current_level,
			current_xp, is_premium, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8
		)
	`

	_, err = s.db.Exec(ctx, query,
		playerBattlePass.ID, playerBattlePass.PlayerID,
		playerBattlePass.BattlePassID, playerBattlePass.CurrentLevel,
		playerBattlePass.CurrentXP, playerBattlePass.IsPremium,
		playerBattlePass.CreatedAt, playerBattlePass.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return playerBattlePass, nil
}

// AddXP adds XP to a player's battle pass progress and handles level ups
func (s *BattlePassService) AddXP(ctx context.Context, playerBattlePassID string, amount int, source string, metadata map[string]interface{}) error {
	if amount <= 0 {
		return ErrInvalidXPAmount
	}

	// Start transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Get current player progress and battle pass info
	query := `
		SELECT pbp.battle_pass_id, pbp.current_level, pbp.current_xp, bp.max_level
		FROM player_battle_passes pbp
		JOIN battle_passes bp ON pbp.battle_pass_id = bp.id
		WHERE pbp.id = $1
		FOR UPDATE
	`

	var currentLevel, currentXP, maxLevel int
	var battlePassID string
	err = tx.QueryRow(ctx, query, playerBattlePassID).Scan(&battlePassID, &currentLevel, &currentXP, &maxLevel)
	if err != nil {
		return err
	}

	// Record XP transaction
	xpTx := &models.BattlePassXPTransaction{
		ID:                 uuid.New().String(),
		PlayerBattlePassID: playerBattlePassID,
		Amount:             amount,
		Source:             source,
		Metadata:           metadata,
		CreatedAt:          time.Now(),
	}

	query = `
		INSERT INTO battle_pass_xp_transactions (
			id, player_battle_pass_id, amount,
			source, metadata, created_at
		) VALUES (
			$1, $2, $3,
			$4, $5, $6
		)
	`

	_, err = tx.Exec(ctx, query,
		xpTx.ID, xpTx.PlayerBattlePassID,
		xpTx.Amount, xpTx.Source,
		xpTx.Metadata, xpTx.CreatedAt,
	)
	if err != nil {
		return err
	}

	// Calculate new level and XP
	newXP := currentXP + amount
	newLevel := currentLevel

	// Check for level ups
	for newLevel < maxLevel {
		// Get next level's required XP
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

	// Update player progress
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

	return tx.Commit(ctx)
}

func (s *BattlePassService) ClaimReward(ctx context.Context, playerBattlePassID string, level int, isPremium bool) (*models.PlayerBattlePassReward, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Check if player has reached the level and has premium if required
	query := `
		SELECT pbp.current_level, pbp.is_premium, bpl.free_rewards, bpl.premium_rewards, pbp.player_id
		FROM player_battle_passes pbp
		JOIN battle_pass_levels bpl ON bpl.battle_pass_id = pbp.battle_pass_id
		WHERE pbp.id = $1 AND bpl.level = $2
		FOR UPDATE
	`

	var currentLevel int
	var hasPremium bool
	var freeRewards, premiumRewards []models.Item
	var playerID string
	err = tx.QueryRow(ctx, query, playerBattlePassID, level).Scan(
		&currentLevel, &hasPremium, &freeRewards, &premiumRewards, &playerID,
	)
	if err != nil {
		return nil, ErrLevelNotFound
	}

	if currentLevel < level {
		return nil, ErrInsufficientLevel
	}

	if isPremium && !hasPremium {
		return nil, ErrPremiumRequired
	}

	// Check if reward was already claimed
	query = `
		SELECT EXISTS (
			SELECT 1
			FROM player_battle_pass_rewards
			WHERE player_battle_pass_id = $1
			AND level = $2
			AND is_premium = $3
		)
	`

	var alreadyClaimed bool
	err = tx.QueryRow(ctx, query, playerBattlePassID, level, isPremium).Scan(&alreadyClaimed)
	if err != nil {
		return nil, err
	}

	if alreadyClaimed {
		return nil, ErrRewardAlreadyClaimed
	}

	claim := &models.PlayerBattlePassReward{
		ID:                 uuid.New().String(),
		PlayerBattlePassID: playerBattlePassID,
		Level:              level,
		IsPremium:          isPremium,
		ClaimedAt:          time.Now(),
		CreatedAt:          time.Now(),
	}

	query = `
		INSERT INTO player_battle_pass_rewards (
			id, player_battle_pass_id, level,
			is_premium, claimed_at, created_at
		) VALUES (
			$1, $2, $3,
			$4, $5, $6
		)
	`

	_, err = tx.Exec(ctx, query,
		claim.ID, claim.PlayerBattlePassID,
		claim.Level, claim.IsPremium,
		claim.ClaimedAt, claim.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Distribute rewards
	rewards := freeRewards
	if isPremium {
		rewards = premiumRewards
	}

	err = s.productService.GiveRewardToPlayer(ctx, rewards, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to give rewards: %w", err)
	}

	return claim, tx.Commit(ctx)
}

// UpgradeToPremium upgrades a player's battle pass to premium
func (s *BattlePassService) UpgradeToPremium(ctx context.Context, playerBattlePassID string) error {
	query := `
		UPDATE player_battle_passes
		SET is_premium = true,
			updated_at = NOW()
		WHERE id = $1
		AND is_premium = false
	`

	result, err := s.db.Exec(ctx, query, playerBattlePassID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("battle pass is already premium or not found")
	}

	return nil
}

// GetPlayerBattlePassProgress gets detailed progress information for a player's battle pass
func (s *BattlePassService) GetPlayerBattlePassProgress(ctx context.Context, playerBattlePassID string) (*models.PlayerBattlePass, error) {
	query := `
		SELECT pbp.id, pbp.player_id, pbp.battle_pass_id,
			   pbp.current_level, pbp.current_xp, pbp.is_premium,
			   pbp.created_at, pbp.updated_at,
			   bp.id, bp.name, bp.description,
			   bp.start_time, bp.end_time, bp.status,
			   bp.max_level, bp.created_at, bp.updated_at
		FROM player_battle_passes pbp
		JOIN battle_passes bp ON pbp.battle_pass_id = bp.id
		WHERE pbp.id = $1
	`

	progress := &models.PlayerBattlePass{
		BattlePass: &models.BattlePass{},
	}

	err := s.db.QueryRow(ctx, query, playerBattlePassID).Scan(
		&progress.ID, &progress.PlayerID, &progress.BattlePassID,
		&progress.CurrentLevel, &progress.CurrentXP, &progress.IsPremium,
		&progress.CreatedAt, &progress.UpdatedAt,
		&progress.BattlePass.ID, &progress.BattlePass.Name,
		&progress.BattlePass.Description, &progress.BattlePass.StartTime,
		&progress.BattlePass.EndTime, &progress.BattlePass.Status,
		&progress.BattlePass.MaxLevel, &progress.BattlePass.CreatedAt,
		&progress.BattlePass.UpdatedAt,
	)

	if err != nil {
		return nil, ErrBattlePassNotFound
	}

	return progress, nil
}

// GetDefaultBattlePassTemplate returns a default template for a battle pass season
func (s *BattlePassService) GetDefaultBattlePassTemplate(maxLevel int) ([]*models.BattlePassLevel, error) {
	if maxLevel <= 0 {
		return nil, errors.New("max level must be greater than 0")
	}

	levels := make([]*models.BattlePassLevel, maxLevel)

	// XP requirement increases with each level
	baseXP := 1000
	xpIncreaseRate := 1.1 // 10% increase per level

	// Rewards get better with level progression
	baseChips := 500
	baseGold := 10
	specialLevels := map[int]bool{
		5: true, 10: true, 15: true, 20: true, 25: true, // Every 5th level
		50: true, 75: true, 100: true, // Milestone levels
	}

	for i := 0; i < maxLevel; i++ {
		level := i + 1
		requiredXP := int(float64(baseXP) * math.Pow(xpIncreaseRate, float64(i)))

		// Basic rewards that increase with level
		freeRewards := []models.Item{
			{Type: models.ItemTypeChips, Amount: baseChips * level},
		}

		premiumRewards := []models.Item{
			{Type: models.ItemTypeChips, Amount: baseChips * level * 2},
			{Type: models.ItemTypeGold, Amount: baseGold + (level / 5)},
		}

		// Add special rewards for milestone levels
		if specialLevels[level] {
			// Special free rewards
			freeRewards = append(freeRewards, models.Item{
				Type:   models.ItemTypeGold,
				Amount: baseGold + (level / 2),
			})

			// Special premium rewards (could be special items, emotes, etc.)
			premiumRewards = append(premiumRewards, models.Item{
				Type:   models.ItemTypeGold,
				Amount: (baseGold + (level / 2)) * 2,
			})
		}

		levels[i] = &models.BattlePassLevel{
			Level:          level,
			RequiredXP:     requiredXP,
			FreeRewards:    freeRewards,
			PremiumRewards: premiumRewards,
		}
	}

	return levels, nil
}

// CreateBattlePassFromTemplate creates a new battle pass season using a template
func (s *BattlePassService) CreateBattlePassFromTemplate(ctx context.Context, name string, description string, startTime time.Time, endTime time.Time, maxLevel int) error {
	// Start transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Create battle pass
	battlePass := &models.BattlePass{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		StartTime:   startTime,
		EndTime:     endTime,
		Status:      models.BattlePassStatusUpcoming,
		MaxLevel:    maxLevel,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	query := `
		INSERT INTO battle_passes (
			id, name, description, start_time, end_time,
			status, max_level, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9
		)
	`

	_, err = tx.Exec(ctx, query,
		battlePass.ID, battlePass.Name, battlePass.Description,
		battlePass.StartTime, battlePass.EndTime,
		battlePass.Status, battlePass.MaxLevel,
		battlePass.CreatedAt, battlePass.UpdatedAt,
	)
	if err != nil {
		return err
	}

	// Get level templates
	levels, err := s.GetDefaultBattlePassTemplate(maxLevel)
	if err != nil {
		return err
	}

	// Create each level
	for _, level := range levels {
		level.ID = uuid.New().String()
		level.BattlePassID = battlePass.ID
		level.CreatedAt = time.Now()

		query = `
			INSERT INTO battle_pass_levels (
				id, battle_pass_id, level, required_xp,
				free_rewards, premium_rewards, created_at
			) VALUES (
				$1, $2, $3, $4,
				$5, $6, $7
			)
		`

		_, err = tx.Exec(ctx, query,
			level.ID, level.BattlePassID, level.Level,
			level.RequiredXP, level.FreeRewards, level.PremiumRewards,
			level.CreatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// GetPlayerClaimedRewards gets all claimed rewards for a player's battle pass
func (s *BattlePassService) GetPlayerClaimedRewards(ctx context.Context, playerBattlePassID string) ([]models.PlayerBattlePassReward, error) {
	query := `
		SELECT id, player_battle_pass_id, level, is_premium, claimed_at, created_at
		FROM player_battle_pass_rewards
		WHERE player_battle_pass_id = $1
		ORDER BY level ASC, is_premium ASC
	`

	rows, err := s.db.Query(ctx, query, playerBattlePassID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rewards := make([]models.PlayerBattlePassReward, 0)
	for rows.Next() {
		var reward models.PlayerBattlePassReward
		err := rows.Scan(
			&reward.ID,
			&reward.PlayerBattlePassID,
			&reward.Level,
			&reward.IsPremium,
			&reward.ClaimedAt,
			&reward.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		rewards = append(rewards, reward)
	}

	return rewards, nil
}

func (s *BattlePassService) GetPlayerBattlePassDetails(ctx context.Context, playerBattlePassID string) (*models.BattlePassProgressDetails, error) {
	// Get player progress
	progress, err := s.GetOrCreatePlayerBattlePass(ctx, playerBattlePassID, playerBattlePassID)
	if err != nil {
		return nil, err
	}

	// Get claimed rewards
	claimedRewards, err := s.GetPlayerClaimedRewards(ctx, playerBattlePassID)
	if err != nil {
		return nil, err
	}

	// Get all levels for this battle pass
	query := `
		SELECT id, battle_pass_id, level, required_xp, free_rewards, premium_rewards, created_at
		FROM battle_pass_levels
		WHERE battle_pass_id = $1
		ORDER BY level ASC
	`

	rows, err := s.db.Query(ctx, query, progress.BattlePassID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var levels []models.BattlePassLevel
	for rows.Next() {
		var level models.BattlePassLevel
		err := rows.Scan(
			&level.ID,
			&level.BattlePassID,
			&level.Level,
			&level.RequiredXP,
			&level.FreeRewards,
			&level.PremiumRewards,
			&level.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		levels = append(levels, level)
	}

	return &models.BattlePassProgressDetails{
		Progress:       *progress,
		ClaimedRewards: claimedRewards,
		Levels:         levels,
	}, nil
}

func (s *BattlePassService) GetOrCreatePlayerBattlePassDetails(ctx context.Context, playerID string) (*models.BattlePassProgressDetails, error) {
	if playerID == "" {
		return nil, ErrPlayerIDRequired
	}

	// Get active battle pass
	battlePass, err := s.GetActiveBattlePass(ctx)
	if err != nil {
		return nil, err
	}

	// Get detailed progress
	details, err := s.GetPlayerBattlePassDetails(ctx, battlePass.ID)
	if err != nil {
		return nil, err
	}

	return details, nil
}
