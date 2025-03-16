package internal

import (
	"encoding/json"
	"errors"
	"sync"

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

type GamePlayerStatus string

const (
	GamePlayerStatusWaiting  GamePlayerStatus = "waiting"
	GamePlayerStatusActive   GamePlayerStatus = "active"
	GamePlayerStatusInactive GamePlayerStatus = "inactive"
	GamePlayerStatusFolded   GamePlayerStatus = "folded"
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

type GameType string

const (
	GameTypeHoldem GameType = "holdem"
)

type IPlayable interface {
	Start() error
	End() error
	OnPlayerJoin(player *GamePlayer) error
	OnPlayerLeave(player *GamePlayer) error
	ProcessAction(action json.RawMessage) error
	DealCards() error
	EvaluateHands() error
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
	GameMessageTypePlayerJoin   GameMessageType = "player_join"
	GameMessageTypePlayerLeave  GameMessageType = "player_leave"
	GameMessageTypePlayerAction GameMessageType = "player_action"
	GameMessageTypeGameStart    GameMessageType = "game_start"
	GameMessageTypeGameEnd      GameMessageType = "game_end"
	GameMessageTypeGameState    GameMessageType = "game_state"
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
	Data       IGamePlayer      `json:"data"`
}

type Game struct {
	ID          string           `json:"id"`
	Status      GameStatus       `json:"status"`
	GameType    GameType         `json:"game_type"`
	Players     []*GamePlayer    `json:"players"`
	Playable    IPlayable        `json:"playable"`
	MinBet      int              `json:"min_bet"`
	MaxPlayers  int              `json:"max_players"`
	ActionChan  chan GameAction  `json:"-"`
	MessageChan chan GameMessage `json:"-"`
	Room        *Room            `json:"-"`
	Mu          sync.RWMutex     `json:"-"`
}

type GameState struct {
	Status   GameStatus    `json:"status"`
	GameType GameType      `json:"game_type"`
	Players  []*GamePlayer `json:"players"`
	State    interface{}   `json:"state"`
}

func NewGame(actionChan chan GameAction, messageChan chan GameMessage, room *Room, maxPlayers int, minBet int, gameType GameType) *Game {
	return &Game{
		ID:          uuid.New().String(),
		Status:      GameStatusWaiting,
		Players:     make([]*GamePlayer, 0),
		MaxPlayers:  maxPlayers,
		MinBet:      minBet,
		ActionChan:  actionChan,
		MessageChan: messageChan,
		Room:        room,
		GameType:    gameType,
		Mu:          sync.RWMutex{},
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
	g.Mu.Lock()
	defer g.Mu.Unlock()

	for _, p := range g.Players {
		if p.Client.User.Player.ID == playerID {
			// g.Players = slices.Delete(g.Players, i, i+1)
			p.Status = GamePlayerStatusInactive
			g.Playable.OnPlayerLeave(p)
			return nil
		}
	}
	return ErrorGamePlayerNotFound
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

func (g *Game) GetGameState() GameState {
	return GameState{
		Status:   g.Status,
		GameType: g.GameType,
		State:    g.Playable.GetGameState(),
	}
}
