package game

import (
	"errors"
	"log"

	"github.com/ahmetkoprulu/rtrp/game/models"
)

// PokerManager handles poker-specific game logic
type PokerManager struct {
	game *models.Game
	deck *models.Deck
}

// NewPokerManager creates a new poker manager for a game
func NewPokerManager(game *models.Game) *PokerManager {
	return &PokerManager{
		game: game,
		deck: models.NewDeck(),
	}
}

// StartNewHand initializes a new hand of poker
func (pm *PokerManager) StartNewHand() error {
	if pm.game.Status != models.GameStatusStarted {
		return errors.New("game not in started state")
	}

	// Reset the deck and shuffle
	pm.deck = models.NewDeck()
	pm.deck.Shuffle()

	// Reset game state for new hand
	pm.game.CommunityCards = make([]models.Card, 0)
	pm.game.Pot = 0
	pm.game.CurrentBet = 0

	// Deal two cards to each active player
	for _, player := range pm.game.Players {
		if !player.Active {
			continue
		}

		player.Cards = make([]models.Card, 0)
		player.Bet = 0
		player.Folded = false
		player.LastAction = ""

		// Deal two cards
		for i := 0; i < 2; i++ {
			card, err := pm.deck.Draw()
			if err != nil {
				return err
			}
			player.Cards = append(player.Cards, card)
		}
	}

	// Move dealer button and set initial positions
	pm.moveDealer()

	log.Printf("[INFO] New hand started - GameID: %s, Players: %d, Dealer: %d",
		pm.game.ID, len(pm.game.Players), pm.game.DealerPosition)

	return nil
}

// DealFlop deals the flop cards
func (pm *PokerManager) DealFlop() error {
	if len(pm.game.CommunityCards) > 0 {
		return errors.New("flop already dealt")
	}

	// Deal three cards for the flop
	for i := 0; i < 3; i++ {
		card, err := pm.deck.Draw()
		if err != nil {
			return err
		}
		card.Hidden = false
		pm.game.CommunityCards = append(pm.game.CommunityCards, card)
	}

	log.Printf("[INFO] Flop dealt - GameID: %s", pm.game.ID)
	return nil
}

// DealTurn deals the turn card
func (pm *PokerManager) DealTurn() error {
	if len(pm.game.CommunityCards) != 3 {
		return errors.New("cannot deal turn: incorrect number of community cards")
	}

	card, err := pm.deck.Draw()
	if err != nil {
		return err
	}
	card.Hidden = false
	pm.game.CommunityCards = append(pm.game.CommunityCards, card)

	log.Printf("[INFO] Turn dealt - GameID: %s", pm.game.ID)
	return nil
}

// DealRiver deals the river card
func (pm *PokerManager) DealRiver() error {
	if len(pm.game.CommunityCards) != 4 {
		return errors.New("cannot deal river: incorrect number of community cards")
	}

	card, err := pm.deck.Draw()
	if err != nil {
		return err
	}
	card.Hidden = false
	pm.game.CommunityCards = append(pm.game.CommunityCards, card)

	log.Printf("[INFO] River dealt - GameID: %s", pm.game.ID)
	return nil
}

// ProcessAction handles a player's action during their turn
func (pm *PokerManager) ProcessAction(action models.GameAction) error {
	player := pm.findPlayer(action.PlayerID)
	if player == nil {
		return errors.New("player not found")
	}

	if player.Folded {
		return errors.New("player has already folded")
	}

	switch action.Action {
	case "fold":
		player.Folded = true
		player.LastAction = "fold"
	case "check":
		if pm.game.CurrentBet > player.Bet {
			return errors.New("cannot check: there is an active bet")
		}
		player.LastAction = "check"
	case "call":
		callAmount := pm.game.CurrentBet - player.Bet
		if callAmount > player.Chips {
			return errors.New("not enough chips to call")
		}
		player.Chips -= callAmount
		player.Bet += callAmount
		pm.game.Pot += callAmount
		player.LastAction = "call"
	case "raise":
		if action.Amount <= pm.game.CurrentBet {
			return errors.New("raise amount must be greater than current bet")
		}
		if action.Amount > player.Chips {
			return errors.New("not enough chips to raise")
		}
		raiseAmount := action.Amount - player.Bet
		player.Chips -= raiseAmount
		player.Bet = action.Amount
		pm.game.CurrentBet = action.Amount
		pm.game.Pot += raiseAmount
		player.LastAction = "raise"
	default:
		return errors.New("invalid action")
	}

	log.Printf("[INFO] Player action processed - GameID: %s, PlayerID: %s, Action: %s, Amount: %d",
		pm.game.ID, player.ID, action.Action, action.Amount)
	return nil
}

// moveDealer advances the dealer position and updates betting positions
func (pm *PokerManager) moveDealer() {
	numPlayers := len(pm.game.Players)
	if numPlayers < 2 {
		return
	}

	// Move dealer button to next active player
	for i := 0; i < numPlayers; i++ {
		nextPos := (pm.game.DealerPosition + 1) % numPlayers
		if pm.game.Players[nextPos].Active {
			pm.game.DealerPosition = nextPos
			break
		}
	}

	// Set initial betting position (player after dealer)
	pm.game.CurrentTurn = (pm.game.DealerPosition + 1) % numPlayers
}

// findPlayer returns a player by their ID
func (pm *PokerManager) findPlayer(playerID string) *models.Player {
	for _, player := range pm.game.Players {
		if player.ID == playerID {
			return player
		}
	}
	return nil
}
