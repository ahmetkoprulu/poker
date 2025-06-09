package services

import (
	"context"
	"fmt"
	"time"

	"github.com/ahmetkoprulu/rtrp/common/data"
	"github.com/ahmetkoprulu/rtrp/models"
	"github.com/google/uuid"
)

type ChallengeService struct {
	db                *data.PgDbContext
	productService    *ProductService    // For handling rewards
	battlePassService *BattlePassService // For handling battle pass XP
}

func NewChallengeService(db *data.PgDbContext, productService *ProductService, battlePassService *BattlePassService) *ChallengeService {
	return &ChallengeService{
		db:                db,
		productService:    productService,
		battlePassService: battlePassService,
	}
}

// CreateChallenge creates a new challenge
func (s *ChallengeService) CreateChallenge(ctx context.Context, challenge *models.Challenge) error {
	challenge.ID = uuid.New().String()
	challenge.CreatedAt = time.Now()

	query := `
		INSERT INTO challenges (
			id, type, category, title, description, 
			requirements, rewards, start_time, end_time, created_at
		) VALUES (
			$1, $2, $3, $4, $5, 
			$6, $7, $8, $9, $10
		)
	`

	_, err := s.db.Exec(ctx, query,
		challenge.ID, challenge.Type, challenge.Category,
		challenge.Title, challenge.Description,
		challenge.Requirements, challenge.Rewards,
		challenge.StartTime, challenge.EndTime, challenge.CreatedAt,
	)

	return err
}

// AssignChallengeToPlayer assigns a challenge to a player
func (s *ChallengeService) AssignChallengeToPlayer(ctx context.Context, playerID string, challengeID string) error {
	playerChallenge := &models.PlayerChallenge{
		ID:          uuid.New().String(),
		PlayerID:    playerID,
		ChallengeID: challengeID,
		Progress:    0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	query := `
		INSERT INTO player_challenges (
			id, player_id, challenge_id, progress, 
			completed, reward_claimed, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, 
			$5, $6, $7, $8
		)
	`

	_, err := s.db.Exec(ctx, query,
		playerChallenge.ID, playerChallenge.PlayerID,
		playerChallenge.ChallengeID, playerChallenge.Progress,
		playerChallenge.Completed, playerChallenge.RewardClaimed,
		playerChallenge.CreatedAt, playerChallenge.UpdatedAt,
	)

	return err
}

// GetPlayerChallenges gets all active challenges for a player
func (s *ChallengeService) GetPlayerChallenges(ctx context.Context, playerID string) ([]*models.PlayerChallenge, error) {
	query := `
		SELECT 
			pc.id, pc.player_id, pc.challenge_id, pc.progress,
			pc.completed, pc.reward_claimed, pc.completed_at,
			pc.created_at, pc.updated_at,
			c.id, c.type, c.category, c.title, c.description,
			c.requirements, c.rewards, c.start_time, c.end_time, c.created_at
		FROM player_challenges pc
		JOIN challenges c ON pc.challenge_id = c.id
		WHERE pc.player_id = $1 AND pc.completed = false
		AND c.end_time > NOW()
	`

	rows, err := s.db.Query(ctx, query, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var challenges []*models.PlayerChallenge
	for rows.Next() {
		pc := &models.PlayerChallenge{
			Challenge: &models.Challenge{},
		}
		err := rows.Scan(
			&pc.ID, &pc.PlayerID, &pc.ChallengeID, &pc.Progress,
			&pc.Completed, &pc.RewardClaimed, &pc.CompletedAt,
			&pc.CreatedAt, &pc.UpdatedAt,
			&pc.Challenge.ID, &pc.Challenge.Type, &pc.Challenge.Category,
			&pc.Challenge.Title, &pc.Challenge.Description,
			&pc.Challenge.Requirements, &pc.Challenge.Rewards,
			&pc.Challenge.StartTime, &pc.Challenge.EndTime,
			&pc.Challenge.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		challenges = append(challenges, pc)
	}

	return challenges, nil
}

// UpdateProgress updates the progress of a player's challenge
func (s *ChallengeService) UpdateProgress(ctx context.Context, playerID string, challengeID string, progress int) error {
	query := `
		UPDATE player_challenges
		SET progress = $1, updated_at = NOW()
		WHERE player_id = $2 AND challenge_id = $3
		RETURNING progress >= (
			SELECT (requirements->>'target_count')::int
			FROM challenges
			WHERE id = $3
		)
	`

	var isCompleted bool
	err := s.db.QueryRow(ctx, query, progress, playerID, challengeID).Scan(&isCompleted)
	if err != nil {
		return err
	}

	// If challenge is completed, mark it as such
	if isCompleted {
		return s.completeChallenge(ctx, playerID, challengeID)
	}

	return nil
}

// completeChallenge marks a challenge as completed
func (s *ChallengeService) completeChallenge(ctx context.Context, playerID string, challengeID string) error {
	now := time.Now()
	query := `
		UPDATE player_challenges
		SET completed = true, completed_at = $1, updated_at = $1
		WHERE player_id = $2 AND challenge_id = $3
	`

	_, err := s.db.Exec(ctx, query, now, playerID, challengeID)
	return err
}

// ClaimRewards claims the rewards for a completed challenge
func (s *ChallengeService) ClaimRewards(ctx context.Context, playerID string, challengeID string) error {
	// Start transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Get challenge and verify it's completed but not claimed
	var rewards []models.Item
	query := `
		UPDATE player_challenges
		SET reward_claimed = true
		WHERE player_id = $1 AND challenge_id = $2
		AND completed = true AND reward_claimed = false
		RETURNING (
			SELECT rewards
			FROM challenges
			WHERE id = $2
		)
	`

	err = tx.QueryRow(ctx, query, playerID, challengeID).Scan(&rewards)
	if err != nil {
		return fmt.Errorf("failed to claim rewards: %w", err)
	}

	// Separate Battle Pass XP rewards from other rewards
	var battlePassXP int
	var otherRewards []models.Item

	for _, reward := range rewards {
		if reward.Type == models.ItemTypeBattlePassXP {
			battlePassXP += reward.Amount
		} else {
			otherRewards = append(otherRewards, reward)
		}
	}

	// Handle non-Battle Pass rewards
	if len(otherRewards) > 0 {
		err = s.productService.GiveRewardToPlayer(ctx, otherRewards, playerID)
		if err != nil {
			return fmt.Errorf("failed to give rewards: %w", err)
		}
	}

	// Handle Battle Pass XP if any
	if battlePassXP > 0 {
		// Get active battle pass
		query = `
			SELECT pbp.id
			FROM player_battle_passes pbp
			JOIN battle_passes bp ON pbp.battle_pass_id = bp.id
			WHERE pbp.player_id = $1
			AND bp.status = 'active'
			AND bp.start_time <= NOW()
			AND bp.end_time >= NOW()
			LIMIT 1
		`

		var playerBattlePassID string
		err = tx.QueryRow(ctx, query, playerID).Scan(&playerBattlePassID)
		if err != nil {
			return fmt.Errorf("failed to get active battle pass: %w", err)
		}

		// Add XP to battle pass
		metadata := map[string]interface{}{
			"source":       "challenge",
			"challenge_id": challengeID,
		}

		err = s.battlePassService.AddXP(ctx, playerBattlePassID, battlePassXP, "challenge", metadata)
		if err != nil {
			return fmt.Errorf("failed to add battle pass XP: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (s *ChallengeService) getDailyChallengeTemplates() []*models.Challenge {
	now := time.Now()
	endTime := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location()).Add(24 * time.Hour)

	return []*models.Challenge{
		{
			Type:        models.ChallengePlayHands,
			Category:    models.CategoryDaily,
			Title:       "Daily Grind",
			Description: "Play 10 hands in poker games",
			Requirements: models.Requirements{
				TargetCount: 10,
			},
			Rewards: []models.Item{
				{Type: models.ItemTypeChips, Amount: 1000},
				{Type: models.ItemTypeBattlePassXP, Amount: 100},
			},
			StartTime: now,
			EndTime:   endTime,
		},
		{
			Type:        models.ChallengeWinHands,
			Category:    models.CategoryDaily,
			Title:       "Winning Streak",
			Description: "Win 3 hands in poker games",
			Requirements: models.Requirements{
				TargetCount: 3,
			},
			Rewards: []models.Item{
				{Type: models.ItemTypeChips, Amount: 2000},
				{Type: models.ItemTypeBattlePassXP, Amount: 200},
			},
			StartTime: now,
			EndTime:   endTime,
		},
		// Add more templates as needed
	}
}

// GenerateDailyChallenges generates and assigns daily challenges for a player
func (s *ChallengeService) GenerateDailyChallenges(ctx context.Context, playerID string) error {
	// First, check if player already has active daily challenges
	query := `
        SELECT COUNT(*) FROM player_challenges pc
        JOIN challenges c ON pc.challenge_id = c.id
        WHERE pc.player_id = $1 AND c.category = $2
        AND c.end_time > NOW() AND pc.completed = false
    `

	var count int
	if err := s.db.QueryRow(ctx, query, playerID, models.CategoryDaily).Scan(&count); err != nil {
		return err
	}

	if count > 0 {
		return nil // Player already has active daily challenges
	}

	// Generate new daily challenges
	templates := s.getDailyChallengeTemplates()

	// Start transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, template := range templates {
		// Create challenge
		if err := s.CreateChallenge(ctx, template); err != nil {
			return err
		}

		// Assign to player
		if err := s.AssignChallengeToPlayer(ctx, playerID, template.ID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// HandleGameEvent processes game events and updates relevant challenges
func (s *ChallengeService) HandleGameEvent(ctx context.Context, playerID string, eventType string, metadata interface{}) error {
	// Get player's active challenges
	challenges, err := s.GetPlayerChallenges(ctx, playerID)
	if err != nil {
		return err
	}

	for _, challenge := range challenges {
		progress := challenge.Progress

		// Update progress based on event type and challenge type
		switch challenge.Challenge.Type {
		case models.ChallengePlayHands:
			if eventType == "hand_completed" {
				progress++
			}
		case models.ChallengeWinHands:
			if eventType == "hand_won" {
				progress++
			}
		case models.ChallengeWinAllIn:
			if eventType == "all_in_won" {
				progress++
			}
		case models.ChallengeGetRoyalFlush:
			if eventType == "hand_completed" {
				// Check if the hand is royal flush from metadata
				if handData, ok := metadata.(map[string]interface{}); ok {
					if handRank, exists := handData["hand_rank"].(string); exists && handRank == "royal_flush" {
						progress++
					}
				}
			}
		}

		// Update progress if changed
		if progress != challenge.Progress {
			if err := s.UpdateProgress(ctx, playerID, challenge.ChallengeID, progress); err != nil {
				return err
			}
		}
	}

	return nil
}

// RefreshDailyChallenges checks and refreshes daily challenges for all players
func (s *ChallengeService) RefreshDailyChallenges(ctx context.Context) error {
	query := `
        UPDATE player_challenges pc
        SET completed = true
        FROM challenges c
        WHERE pc.challenge_id = c.id
        AND c.category = $1
        AND c.end_time < NOW()
        AND pc.completed = false
    `

	_, err := s.db.Exec(ctx, query, models.CategoryDaily)
	return err
}

// func (s *ChallengeService) validateChallenge(challenge *models.Challenge) error {
// 	if challenge.Type == "" || challenge.Category == "" {
// 		return fmt.Errorf("challenge type and category are required")
// 	}

// 	if challenge.Title == "" || challenge.Description == "" {
// 		return fmt.Errorf("challenge title and description are required")
// 	}

// 	if challenge.Requirements.TargetCount <= 0 {
// 		return fmt.Errorf("target count must be greater than 0")
// 	}

// 	if len(challenge.Rewards) == 0 {
// 		return fmt.Errorf("challenge must have at least one reward")
// 	}

// 	return nil
// }

// // isValidEventForChallenge checks if an event is valid for a challenge type
// func (s *ChallengeService) isValidEventForChallenge(eventType string, challengeType models.ChallengeType) bool {
// 	validEvents := map[models.ChallengeType][]string{
// 		models.ChallengePlayHands:     {"hand_completed"},
// 		models.ChallengeWinHands:      {"hand_won"},
// 		models.ChallengeWinAllIn:      {"all_in_won"},
// 		models.ChallengeGetRoyalFlush: {"hand_completed"},
// 		models.ChallengePlayGames:     {"game_completed"},
// 	}

// 	events, exists := validEvents[challengeType]
// 	if !exists {
// 		return false
// 	}

// 	for _, e := range events {
// 		if e == eventType {
// 			return true
// 		}
// 	}
// 	return false
// }
