package api

import (
	"context"
	"fmt"
	"time"
)

type PlayerService struct {
	parent   *ApiService
	endpoint string
	client   *ApiClient
}

func NewPlayerService(parent *ApiService, endpoint string) *PlayerService {
	service := &PlayerService{
		parent:   parent,
		endpoint: parent.config.BaseURL + endpoint,
	}
	service.client = parent.getClient("player-service", endpoint)

	return service
}

func (s *PlayerService) UpdateChip(playerID string, chips int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var response ApiResponse[int]

	err := s.client.Put(ctx, s.endpoint+"/chip", IncrementChipRequest{
		ID:     playerID,
		Amount: chips,
	}, &response)
	if err != nil {
		return 0, fmt.Errorf("failed to update chip: %w", err)
	}

	return response.Data, nil
}

func (s *PlayerService) UpdateChips(players []IncrementChipRequest) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var response ApiResponse[int]

	err := s.client.Put(ctx, s.endpoint+"/chips", IncrementChipsRequest{
		Players: players,
	}, &response)
	if err != nil {
		return 0, fmt.Errorf("failed to update chips: %w", err)
	}

	return response.Data, nil
}

type IncrementChipRequest struct {
	ID     string `json:"id"`
	Amount int    `json:"amount"`
}

type IncrementChipsRequest struct {
	Players []IncrementChipRequest `json:"players"`
}
