package internal

import (
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/ahmetkoprulu/rtrp/game/internal/mq"
	"github.com/ahmetkoprulu/rtrp/game/models"
	"github.com/google/uuid"
)

type GameStatus string

const (
	GameStatusWaiting  GameStatus = "waiting"
	GameStatusStarting GameStatus = "starting"
	GameStatusStarted  GameStatus = "started"
	GameStatusEnding   GameStatus = "ending"
	GameStatusEnd      GameStatus = "end"
)

type GamePlayerStatus int

const (
	GamePlayerStatusWaiting GamePlayerStatus = iota
	GamePlayerStatusActive
	GamePlayerStatusInactive
	GamePlayerStatusFolded
)

type GameActionType string

const (
	GameActionTypePlayerJoin   GameActionType = "player_join"
	GameActionTypePlayerLeave  GameActionType = "player_leave"
	GameActionTypePlayerAction GameActionType = "player_action"
)

type GameError error

var (
	ErrorGameFull            GameError = errors.New("game_full")
	ErrorGamePlayerAlreadyIn GameError = errors.New("game_player_already_in")
	ErrorGamePositionTaken   GameError = errors.New("game_position_taken")
	ErrorGamePlayerNotFound  GameError = errors.New("game_player_not_found")
	ErrorGameNotReady        GameError = errors.New("game_not_ready")
)

type GameType int

const (
	GameTypeHoldem GameType = 1
)

type IPlayable interface {
	Start() error
	End() error
	OnPlayerJoin(player *GamePlayer) error
	OnPlayerLeave(player *GamePlayer) error
	ProcessAction(action json.RawMessage) error
	DealCards() error
	CanStart() bool
	GetGameState() interface{}
}

type GameAction struct {
	PlayerID   string          `json:"player_id"`
	ActionType GameActionType  `json:"action_type"`
	Data       json.RawMessage `json:"data"`
}

type GameMessage struct {
	PlayerID    string          `json:"player_id"`
	ToGame      bool            `json:"to_game"`
	ToRoom      bool            `json:"to_room"`
	MessageType GameMessageType `json:"message_type"`
	Data        interface{}     `json:"data"`
}

type GameMessageType string

const (
	GameMessageTypePlayerAction GameMessageType = "player_action"
)

type IGamePlayer interface {
	GetBalance() int
	GetData() interface{}
}

type GamePlayer struct {
	Position   int              `json:"position"`
	Balance    int              `json:"balance"`
	LastAction string           `json:"last_action"`
	Client     *Client          `json:"client"`
	Status     GamePlayerStatus `json:"status"`
}

type Game struct {
	ID          string               `json:"id"`
	Status      GameStatus           `json:"status"`
	GameType    GameType             `json:"game_type"`
	Players     []*GamePlayer        `json:"players"`
	Playable    IPlayable            `json:"playable"`
	MinBet      int                  `json:"min_bet"`
	MaxPlayers  int                  `json:"max_players"`
	ActionChan  chan GameAction      `json:"-"`
	MessageChan chan models.Response `json:"-"`
	Room        *Room                `json:"-"`
	Mu          sync.RWMutex         `json:"-"`

	GameEventPublisher *mq.GameEventPublisher
}

func NewGame(actionChan chan GameAction, messageChan chan models.Response, room *Room, maxPlayers int, minBet int, gameType GameType) *Game {
	gameEventPublisher, err := mq.NewGameEventPublisher()
	if err != nil {
		log.Printf("Error creating game event publisher: %v", err)
	}

	return &Game{
		ID:                 uuid.New().String(),
		Status:             GameStatusWaiting,
		Players:            make([]*GamePlayer, 0),
		MaxPlayers:         maxPlayers,
		MinBet:             minBet,
		ActionChan:         actionChan,
		MessageChan:        messageChan,
		Room:               room,
		GameType:           gameType,
		Mu:                 sync.RWMutex{},
		GameEventPublisher: gameEventPublisher,
	}
}

func (g *Game) AddPlayer(position int, player *Client) error {
	g.Mu.Lock()
	defer g.Mu.Unlock()

	gamePlayer := &GamePlayer{
		Position: position,
		Client:   player,
		Balance:  int(player.User.Player.Chips),
	}

	if len(g.Players) >= g.MaxPlayers {
		return ErrorGameFull
	}

	for _, p := range g.Players {
		if p.Client.User.Player.ID == player.User.Player.ID {
			return ErrorGamePlayerAlreadyIn
		}

		if p.Position == position {
			return ErrorGamePositionTaken
		}
	}

	gamePlayer.Status = GamePlayerStatusWaiting
	g.Players = append(g.Players, gamePlayer)
	player.CurrentGame = g
	g.Playable.OnPlayerJoin(gamePlayer)

	return nil
}

func (g *Game) RemovePlayer(playerID string) error {
	// g.Mu.Lock()
	// defer g.Mu.Unlock()

	for _, p := range g.Players {
		if p.Client.User.Player.ID == playerID {
			// g.Players = slices.Delete(g.Players, i, i+1)
			p.Status = GamePlayerStatusInactive
			g.Playable.OnPlayerLeave(p)
			return nil
		}
	}
	return nil
}

func (g *Game) Start() error {
	if !g.Playable.CanStart() {
		return ErrorGameNotReady
	}

	err := g.Playable.Start()
	if err != nil {
		return err
	}

	g.Status = GameStatusStarted
	return nil
}

func (g *Game) End() error {
	err := g.Playable.End()
	if err != nil {
		return err
	}

	g.Status = GameStatusEnd
	return nil
}

func (g *Game) GetGameState() interface{} {
	return g.Playable.GetGameState()
}

// Reset clears all players and resets game state
func (g *Game) Reset() error {
	g.Mu.Lock()
	defer g.Mu.Unlock()

	log.Printf("[ADMIN] Resetting game %s\n", g.ID)

	// Stop the game if it's running
	if g.Status == GameStatusStarted || g.Status == GameStatusStarting {
		if g.Playable != nil {
			err := g.Playable.End()
			if err != nil {
				log.Printf("[ADMIN] Error ending game: %v\n", err)
			}
		}
	}

	// Clear all players
	for _, player := range g.Players {
		player.Client.CurrentGame = nil
		player.Status = GamePlayerStatusInactive
	}
	g.Players = make([]*GamePlayer, 0)

	// Reset game status
	g.Status = GameStatusWaiting

	// Reset playable if it's Holdem
	if holdem, ok := g.Playable.(*Holdem); ok {
		holdem.RefreshState()
		// Send done signal to stop any running goroutines
		select {
		case holdem.doneChannel <- true:
		default:
			// Channel might be full or closed, ignore
		}
	}

	log.Printf("[ADMIN] Game %s reset complete\n", g.ID)
	return nil
}

func (g *Game) UpdatePlayerChips(playerChanges []mq.PlayerChipChange) error {
	chipUpdate := &mq.ChipUpdateMessage{
		MessageID:     uuid.New().String(),
		RoomID:        g.Room.ID,
		GameType:      int(g.GameType),
		PlayerChanges: playerChanges,
		Timestamp:     time.Now(),
	}

	err := g.GameEventPublisher.PublishChipUpdate(chipUpdate)
	if err != nil {
		log.Printf("[ERROR] Failed to publish chip update: %v", err)
	}

	return err
}
