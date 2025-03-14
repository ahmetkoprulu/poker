package services

import (
	"context"
	"errors"
	"math/rand"

	"github.com/ahmetkoprulu/rtrp/internal/config"
	"github.com/ahmetkoprulu/rtrp/models"
)

type MiniGameService struct {
	playerService  *PlayerService
	productService *ProductService
}

func NewMiniGameService(playerService *PlayerService, productService *ProductService) *MiniGameService {
	return &MiniGameService{
		playerService:  playerService,
		productService: productService,
	}
}

func (s *MiniGameService) SpinSlot(ctx context.Context, playerID string) (*models.Item, []models.ItemType, error) {
	player, err := s.playerService.GetPlayerByID(ctx, playerID)
	if err != nil {
		return nil, nil, err
	}

	if player == nil {
		return nil, nil, errors.New("player_not_found")
	}

	if player.MiniGames.Slots <= 0 {
		return nil, nil, errors.New("insufficient_slots")
	}

	config, err := config.GetGameConfig()
	if err != nil {
		return nil, nil, err
	}

	rewards := config.MiniGames.SlotRewards
	indices := make([]models.ItemType, 0, 3)
	spinRewards := make([]models.Item, 0, 3)
	for i := 0; i < 3; i++ {
		index := rand.Intn(len(rewards))
		indices = append(indices, rewards[index].Type)
		spinRewards = append(spinRewards, rewards[index])
	}

	reward := s.checkSlotReward(spinRewards)
	err = s.playerService.DecrementMiniGamePoints(ctx, playerID, "slots", "last_slot_played_at")
	if err != nil {
		return nil, nil, err
	}

	err = s.productService.GiveRewardToPlayer(ctx, []models.Item{reward}, playerID)
	if err != nil {
		return nil, nil, err
	}

	return &reward, indices, nil
}

func (s *MiniGameService) SpinWheel(ctx context.Context, playerID string) (int, *models.Item, error) {
	player, err := s.playerService.GetPlayerByID(ctx, playerID)
	if err != nil {
		return 0, nil, err
	}

	if player.MiniGames.Wheels <= 0 {
		return 0, nil, errors.New("insufficient_wheels")
	}

	config, err := config.GetGameConfig()
	if err != nil {
		return 0, nil, err
	}

	index := rand.Intn(len(config.MiniGames.WheelRewards))
	reward := config.MiniGames.WheelRewards[index]

	err = s.playerService.DecrementMiniGamePoints(ctx, playerID, "wheels", "last_wheel_played_at")
	if err != nil {
		return 0, nil, err
	}

	err = s.productService.GiveRewardToPlayer(ctx, []models.Item{reward}, playerID)
	if err != nil {
		return 0, nil, err
	}

	return index, &reward, nil
}

func (s *MiniGameService) SpinGoldWheel(ctx context.Context, playerID string) (int, *models.Item, error) {
	player, err := s.playerService.GetPlayerByID(ctx, playerID)
	if err != nil {
		return 0, nil, err
	}

	if player.MiniGames.GoldWheels <= 0 {
		return 0, nil, errors.New("insufficient_gold_wheels")
	}

	config, err := config.GetGameConfig()
	if err != nil {
		return 0, nil, err
	}

	index := rand.Intn(len(config.MiniGames.GoldWheelRewards))
	reward := config.MiniGames.GoldWheelRewards[index]

	err = s.playerService.DecrementMiniGamePoints(ctx, playerID, "gold_wheels", "last_wheel_played_at")
	if err != nil {
		return 0, nil, err
	}

	err = s.productService.GiveRewardToPlayer(ctx, []models.Item{reward}, playerID)
	if err != nil {
		return 0, nil, err
	}

	return index, &reward, nil
}

func (s *MiniGameService) checkSlotReward(spinRewards []models.Item) models.Item {
	if spinRewards[0].Type == spinRewards[1].Type && spinRewards[1].Type == spinRewards[2].Type {
		spinRewards[0].Amount *= 5
		return spinRewards[0]
	}

	if spinRewards[0].Type == spinRewards[1].Type || spinRewards[0].Type == spinRewards[2].Type {
		spinRewards[0].Amount *= 2
		return spinRewards[0]
	}

	if spinRewards[1].Type == spinRewards[2].Type {
		spinRewards[1].Amount *= 2
		return spinRewards[1]
	}

	return models.Item{
		Type:   models.ItemTypeChips,
		Amount: 1000000,
	}
}
