package services

import (
	"context"
	"fmt"

	"github.com/ahmetkoprulu/rtrp/common/data"
	"github.com/ahmetkoprulu/rtrp/models"
)

type LobbyService struct {
	db                *data.PgDbContext
	playerService     *PlayerService
	eventService      *EventService
	battlePassService *BattlePassService
}

func NewLobbyService(db *data.PgDbContext, playerService *PlayerService, eventService *EventService, battlePassService *BattlePassService) *LobbyService {
	return &LobbyService{db: db, playerService: playerService, eventService: eventService, battlePassService: battlePassService}
}

func (s *LobbyService) GetState(ctx context.Context, playerID string) (*LobbyState, error) {
	if playerID == "" {
		return nil, fmt.Errorf("player id is required")
	}

	playerEventSchedules, err := s.eventService.GetPlayerEventSchedules(ctx, playerID)
	if err != nil {
		return nil, err
	}

	playerBattlePass, err := s.battlePassService.GetOrCreatePlayerBattlePassDetails(context.Background(), playerID)
	if err != nil && err != ErrBattlePassNotFound {
		return nil, err
	}

	return &LobbyState{
		Events:     playerEventSchedules,
		BattlePass: playerBattlePass,
	}, nil
}

type LobbyState struct {
	Player     *models.Player                      `json:"player"`
	Events     []*models.PlayerEventScheduleDetail `json:"events"`
	BattlePass *models.BattlePassProgressDetails   `json:"battle_pass"`
}
