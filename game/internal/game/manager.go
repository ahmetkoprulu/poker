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
	games map[string]*models.Game
	mu    sync.RWMutex
}

// NewGameManager creates a new game manager instance
func NewGameManager() *GameManager {
	return &GameManager{
		games: make(map[string]*models.Game),
	}
}

// CreateGame creates a new game instance
func (gm *GameManager) CreateGame(maxPlayers, minBet int) *models.Game {
	game := models.NewGame(maxPlayers, minBet)

	gm.mu.Lock()
	gm.games[game.ID] = game
	gm.mu.Unlock()

	return game
}

// RegisterGame registers an existing game with the manager
func (gm *GameManager) RegisterGame(game *models.Game) {
	if game == nil {
		log.Printf("[ERROR] Attempted to register nil game")
		return
	}

	gm.mu.Lock()
	gm.games[game.ID] = game
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
	game.DealerPosition = 0 // Start with first player as dealer
	game.CurrentTurn = 1    // Start with player after dealer
	game.CurrentBet = 0
	game.Pot = 0

	// Reset player states
	for _, player := range game.Players {
		player.Bet = 0
		player.Folded = false
		player.Active = true
		player.LastAction = ""
	}

	log.Printf("[INFO] Game started successfully - GameID: %s, PlayerCount: %d, Dealer: %d, FirstToAct: %d",
		gameID, len(game.Players), game.DealerPosition, game.CurrentTurn)
	return nil
}

// RemoveGame removes a game from the manager
func (gm *GameManager) RemoveGame(gameID string) {
	gm.mu.Lock()
	delete(gm.games, gameID)
	gm.mu.Unlock()
}
