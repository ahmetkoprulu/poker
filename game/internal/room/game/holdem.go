package game

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/ahmetkoprulu/rtrp/game/models"
)

type Holdem struct {
	State          HoldemState
	deck           *models.Deck
	game           *models.Game
	actionsChannel chan models.GameAction
}

type HoldemState struct {
	CommunityCards []models.Card
	Pot            int
	CurrentBet     int
	DealerPosition int
	CurrentTurn    int
}

type HoldemAction struct {
	PlayerID string
	Action   string
	Amount   int
}

func NewHoldem(game *models.Game) *Holdem {
	return &Holdem{
		State:          HoldemState{},
		actionsChannel: game.ActionChan,
		deck:           models.NewDeck(),
		game:           game,
	}
}

func (h *Holdem) Start() error {
	h.State = HoldemState{}
	h.deck = models.NewDeck()
	h.game.Status = models.GameStatusStarted

	return nil
}

func (h *Holdem) End() error {
	h.State = HoldemState{}
	h.deck = models.NewDeck()
	h.game.Status = models.GameStatusFinished
	return nil
}

func (h *Holdem) ProcessAction(msg json.RawMessage) error {
	var action HoldemAction
	if err := json.Unmarshal(msg, &action); err != nil {
		return err
	}

	player := h.game.Players[h.State.CurrentTurn]
	if action.PlayerID != player.Player.ID {
		return errors.New("action received from wrong player")
	}

	if player == nil {
		return errors.New("player not found")
	}

	switch action.Action {
	case "fold":
		player.Hand = nil
	case "check":
		if h.State.CurrentBet > 0 {
			return errors.New("cannot check, must call or raise")
		}
	case "call":
		callAmount := h.State.CurrentBet - player.Balance
		if callAmount > player.Balance {
			return errors.New("insufficient balance to call")
		}
		player.Balance -= callAmount
		h.State.Pot += callAmount
	case "bet":
		if action.Amount <= h.State.CurrentBet {
			return errors.New("bet must be greater than current bet")
		}
		if action.Amount > player.Balance {
			return errors.New("insufficient balance to bet")
		}
		player.Balance -= action.Amount
		h.State.Pot += action.Amount
		h.State.CurrentBet = action.Amount
	case "raise":
		raiseAmount := action.Amount - h.State.CurrentBet
		if raiseAmount <= 0 {
			return errors.New("raise must be greater than current bet")
		}
		if raiseAmount > player.Balance {
			return errors.New("insufficient balance to raise")
		}
		player.Balance -= raiseAmount
		h.State.Pot += raiseAmount
		h.State.CurrentBet = action.Amount
	default:
		return errors.New("invalid action")
	}

	h.State.CurrentTurn = (h.State.CurrentTurn + 1) % len(h.game.Players)

	return nil
}

func (h *Holdem) CanStart() bool {
	activePlayers := 0
	for _, player := range h.game.Players {
		if player.Status == models.GamePlayerStatusActive || player.Status == models.GamePlayerStatusWaiting {
			activePlayers++
		}
	}

	return activePlayers >= 2
}

func (h *Holdem) DealCards() error {
	h.deck.Shuffle()

	// Deal two cards to each player
	for _, player := range h.game.Players {
		card1, err := h.deck.Draw()
		if err != nil {
			return err
		}
		card2, err := h.deck.Draw()
		if err != nil {
			return err
		}
		player.Hand = append(player.Hand, card1, card2)
	}

	// Set up the community cards (flop, turn, river)
	communityCards := make([]models.Card, 0, 5)
	for i := 0; i < 5; i++ {
		card, err := h.deck.Draw()
		if err != nil {
			return err
		}
		communityCards = append(communityCards, card)
	}
	h.State.CommunityCards = communityCards

	return nil
}

func (h *Holdem) EvaluateHands() error {
	// Evaluate players' hands and determine the winner
	return nil
}

func (h *Holdem) GetGameState() interface{} {
	return h.State
}

func (h *Holdem) PlayRound() error {
	h.HandlePlayers()

	// Pre-Flop: Deal hole cards and start the first betting round
	err := h.DealCards()
	if err != nil {
		return err
	}
	h.State.CurrentBet = 0
	h.State.Pot = 0

	// Flop: Deal three community cards and start the second betting round
	h.State.CommunityCards = h.State.CommunityCards[:3]
	err = h.BettingRound()
	if err != nil {
		return err
	}

	// Turn: Deal the fourth community card and start the third betting round
	h.State.CommunityCards = h.State.CommunityCards[:4]
	err = h.BettingRound()
	if err != nil {
		return err
	}

	// River: Deal the fifth community card and start the final betting round
	h.State.CommunityCards = h.State.CommunityCards[:5]
	err = h.BettingRound()
	if err != nil {
		return err
	}

	// Showdown: Evaluate hands and determine the winner
	err = h.EvaluateHands()
	if err != nil {
		return err
	}

	return nil
}

func (h *Holdem) BettingRound() error {
	for i := 0; i < len(h.game.Players); i++ {

		timer := time.NewTimer(10 * time.Second)
		defer timer.Stop()

		var action models.GameAction
		select {
		case action = <-h.actionsChannel:
			if err := h.ProcessAction(action.Data); err != nil {
				return err
			}
		case <-timer.C:
			player := h.game.Players[h.State.CurrentTurn]
			fold := models.GameAction{PlayerID: player.Player.ID, Action: "fold"}
			data, err := json.Marshal(fold)
			if err != nil {
				return err
			}
			if err := h.ProcessAction(data); err != nil {
				return err
			}
		}
	}

	return nil
}

func (h *Holdem) HandlePlayers() {
	for _, player := range h.game.Players {
		if player.Status == models.GamePlayerStatusWaiting {
			player.Status = models.GamePlayerStatusActive
		}

		if player.Status == models.GamePlayerStatusInactive {
			h.game.RemovePlayer(player.Player.ID)
		}
	}
}
