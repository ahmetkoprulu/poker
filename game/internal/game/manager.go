package game

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/ahmetkoprulu/rtrp/game/models"
)

// GameManager handles all active games and their states
type GameManager struct {
	games         map[string]*models.Game
	pokerManagers map[string]*PokerManager
	mu            sync.RWMutex
}

// NewGameManager creates a new game manager instance
func NewGameManager() *GameManager {
	return &GameManager{
		games:         make(map[string]*models.Game),
		pokerManagers: make(map[string]*PokerManager),
	}
}

// CreateGame creates a new game instance
func (gm *GameManager) CreateGame(maxPlayers, minBet int) *models.Game {
	game := models.NewGame(maxPlayers, minBet)
	pokerManager := NewPokerManager(game)

	gm.mu.Lock()
	gm.games[game.ID] = game
	gm.pokerManagers[game.ID] = pokerManager
	gm.mu.Unlock()

	return game
}

// RegisterGame registers an existing game with the manager
func (gm *GameManager) RegisterGame(game *models.Game) {
	if game == nil {
		log.Printf("[ERROR] Attempted to register nil game")
		return
	}

	pokerManager := NewPokerManager(game)

	gm.mu.Lock()
	gm.games[game.ID] = game
	gm.pokerManagers[game.ID] = pokerManager
	gm.mu.Unlock()

	log.Printf("[INFO] Game registered - GameID: %s, MaxPlayers: %d, MinBet: %d",
		game.ID, game.MaxPlayers, game.MinBet)
}

// GetGame retrieves a game by its ID
func (gm *GameManager) GetGame(gameID string) (*models.Game, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	game, exists := gm.games[gameID]
	if !exists {
		return nil, errors.New("game not found")
	}

	return game, nil
}

// JoinGame adds a player to a game
func (gm *GameManager) JoinGame(gameID string, player *models.Player) error {
	game, err := gm.GetGame(gameID)
	if err != nil {
		log.Printf("[ERROR] Game not found for join - GameID: %s, PlayerID: %s", gameID, player.ID)
		return err
	}

	log.Printf("[INFO] Attempting to add player to game - GameID: %s, PlayerID: %s, CurrentPlayers: %d",
		gameID, player.ID, len(game.Players))

	if !game.AddPlayer(player) {
		log.Printf("[ERROR] Game is full - GameID: %s, PlayerID: %s, MaxPlayers: %d",
			gameID, player.ID, game.MaxPlayers)
		return errors.New("cannot join game: game is full")
	}

	log.Printf("[INFO] Player added to game - GameID: %s, PlayerID: %s, TotalPlayers: %d",
		gameID, player.ID, len(game.Players))
	return nil
}

// LeaveGame removes a player from a game
func (gm *GameManager) LeaveGame(gameID string, playerID string) error {
	game, err := gm.GetGame(gameID)
	if err != nil {
		log.Printf("[ERROR] Game not found for leave - GameID: %s, PlayerID: %s", gameID, playerID)
		return err
	}

	log.Printf("[INFO] Player leaving game - GameID: %s, PlayerID: %s, CurrentPlayers: %d",
		gameID, playerID, len(game.Players))

	// Find the player and mark them as inactive before removing
	playerFound := false
	for _, p := range game.Players {
		if p.ID == playerID {
			p.Active = false
			p.Folded = true
			playerFound = true
			break
		}
	}

	if !playerFound {
		log.Printf("[ERROR] Player not found in game - GameID: %s, PlayerID: %s", gameID, playerID)
		return fmt.Errorf("player not found in game")
	}

	if !game.RemovePlayer(playerID) {
		log.Printf("[ERROR] Failed to remove player - GameID: %s, PlayerID: %s", gameID, playerID)
		return errors.New("failed to remove player from game")
	}

	// If game is active and not enough players, end the game
	if game.Status == models.GameStatusStarted && len(game.Players) < 2 {
		log.Printf("[INFO] Ending game due to insufficient players - GameID: %s, RemainingPlayers: %d",
			gameID, len(game.Players))
		game.Status = models.GameStatusFinished
	}

	log.Printf("[INFO] Player removed from game - GameID: %s, PlayerID: %s, RemainingPlayers: %d",
		gameID, playerID, len(game.Players))
	return nil
}

// StartGame starts a game if possible
func (gm *GameManager) StartGame(gameID string) error {
	game, err := gm.GetGame(gameID)
	if err != nil {
		log.Printf("[ERROR] Game not found for start - GameID: %s", gameID)
		return err
	}

	log.Printf("[INFO] Attempting to start game - GameID: %s, PlayerCount: %d, Status: %s",
		gameID, len(game.Players), game.Status)

	if !game.CanStart() {
		log.Printf("[ERROR] Cannot start game - GameID: %s, PlayerCount: %d, Status: %s",
			gameID, len(game.Players), game.Status)
		return errors.New("cannot start game: not enough players or invalid state")
	}

	// Initialize game state
	game.Status = models.GameStatusStarted

	// Start the first hand
	pokerManager := gm.pokerManagers[gameID]
	if err := pokerManager.StartNewHand(); err != nil {
		log.Printf("[ERROR] Failed to start first hand - GameID: %s, Error: %s", gameID, err)
		return fmt.Errorf("failed to start first hand: %w", err)
	}

	log.Printf("[INFO] Game started successfully - GameID: %s, PlayerCount: %d",
		gameID, len(game.Players))
	return nil
}

// ProcessAction processes a player's action in a game
func (gm *GameManager) ProcessAction(gameID string, action models.GameAction) error {
	game, err := gm.GetGame(gameID)
	if err != nil {
		return err
	}

	if game.Status != models.GameStatusStarted {
		return errors.New("game not in progress")
	}

	pokerManager := gm.pokerManagers[gameID]
	return pokerManager.ProcessAction(action)
}

// DealNextRound deals the next round of community cards
func (gm *GameManager) DealNextRound(gameID string) error {
	game, err := gm.GetGame(gameID)
	if err != nil {
		return err
	}

	if game.Status != models.GameStatusStarted {
		return errors.New("game not in progress")
	}

	pokerManager := gm.pokerManagers[gameID]
	communityCards := len(game.CommunityCards)

	switch communityCards {
	case 0:
		return pokerManager.DealFlop()
	case 3:
		return pokerManager.DealTurn()
	case 4:
		return pokerManager.DealRiver()
	default:
		return fmt.Errorf("invalid number of community cards: %d", communityCards)
	}
}

// RemoveGame removes a game from the manager
func (gm *GameManager) RemoveGame(gameID string) {
	gm.mu.Lock()
	delete(gm.games, gameID)
	delete(gm.pokerManagers, gameID)
	gm.mu.Unlock()
}
