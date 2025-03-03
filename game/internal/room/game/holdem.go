package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/ahmetkoprulu/rtrp/game/models"
)

type HoldemActionType string

const (
	HoldemActionStart      HoldemActionType = "join"
	HoldemActionFold       HoldemActionType = "fold"
	HoldemActionCall       HoldemActionType = "call"
	HoldemActionRaise      HoldemActionType = "raise"
	HoldemActionBet        HoldemActionType = "bet"
	HoldemActionCheck      HoldemActionType = "check"
	HoldemActionAllIn      HoldemActionType = "allIn"
	HoldemActionSmallBlind HoldemActionType = "smallBlind"
	HoldemActionBigBlind   HoldemActionType = "bigBlind"
	HoldemActionNewRound   HoldemActionType = "newRound"
)

type Holdem struct {
	State          HoldemState
	deck           *models.Deck
	game           *models.Game
	actionsChannel chan models.GameAction
}

type HoldemAction struct {
	PlayerID string           `json:"playerId"`
	Action   HoldemActionType `json:"action"`
	Amount   int              `json:"amount"`
}

type HoldemState struct {
	CommunityCards []models.Card
	Pot            int

	CurrentRound HoldemRound
	CurrentTurn  int
	CurrentBet   int

	DealerIndex      int
	SmallBlindAmount int
	SmallBlindIndex  int
	BigBlindAmount   int
	BigBlindIndex    int

	RoundComplete     bool
	LastRaisePosition int
	PlayerBets        map[string]int
	PlayerHands       map[string][]models.Card

	// Side pots for all-in situations
	// SidePots []SidePot
}

type HoldemPlayer struct {
	PlayerID string
	Balance  int
	Hand     []models.Card
}

func (h HoldemPlayer) GetBalance() int {
	return h.Balance
}

func (h HoldemPlayer) GetData() interface{} {
	return h
}

type HoldemRound int

const (
	PreFlop HoldemRound = iota
	Flop
	Turn
	River
	Showdown
)

func NewHoldem(game *models.Game) *Holdem {
	return &Holdem{
		State: HoldemState{
			SmallBlindAmount:  5,  // Default small blind
			BigBlindAmount:    10, // Default big blind
			CurrentRound:      PreFlop,
			Pot:               0,
			CurrentBet:        0,
			DealerIndex:       -1,
			CurrentTurn:       0,
			SmallBlindIndex:   0,
			BigBlindIndex:     0,
			RoundComplete:     false,
			LastRaisePosition: 0,
			PlayerBets:        make(map[string]int),
			PlayerHands:       make(map[string][]models.Card),
		},
		actionsChannel: game.ActionChan,
		deck:           models.NewDeck(),
		game:           game,
	}
}

func (h *Holdem) ProcessAction(msg json.RawMessage) error {
	var action HoldemAction
	if err := json.Unmarshal(msg, &action); err != nil {
		return err
	}

	player := h.game.Players[h.State.CurrentTurn]
	log.Printf("[INFO] Player %s processing action: %+v", player.Player.ID, action)
	if action.PlayerID != player.Player.ID {
		return errors.New("action received from wrong player")
	}

	if action.Action == HoldemActionRaise && h.State.CurrentBet == 0 { // Automatically convert raise to bet when no current bet exists
		action.Action = HoldemActionBet
	} else if action.Action == HoldemActionBet && h.State.CurrentBet > 0 {
		action.Action = HoldemActionRaise
	}

	playerBet := h.State.PlayerBets[player.Player.ID]
	toCall := min(h.State.CurrentBet-playerBet, player.Balance)

	switch action.Action {
	case HoldemActionFold:
		log.Printf("[INFO] Player %s folds", player.Player.ID)
		h.State.PlayerHands[player.Player.ID] = nil

	case HoldemActionCheck:
		if toCall > 0 {
			return errors.New("cannot check, must call or raise")
		}
		log.Printf("[INFO] Player %s checks", player.Player.ID)

	case HoldemActionCall:
		if toCall == 0 {
			return errors.New("nothing to call, must check")
		}

		player.Balance -= toCall
		h.State.Pot += toCall
		h.State.PlayerBets[player.Player.ID] += toCall

		log.Printf("[INFO] Player %s calls %d", player.Player.ID, toCall)

		// If player is all-in, log it
		if player.Balance == 0 {
			log.Printf("[INFO] Player %s is all-in", player.Player.ID)
		}

	case HoldemActionBet:
		if h.State.CurrentBet > 0 {
			return errors.New("cannot bet, must raise")
		}

		// Validate bet amount
		minBet := h.State.BigBlindAmount
		if action.Amount < minBet {
			return errors.New("bet must be at least the big blind")
		}

		if action.Amount > player.Balance {
			return errors.New("insufficient balance to bet")
		}

		player.Balance -= action.Amount
		h.State.Pot += action.Amount
		h.State.PlayerBets[player.Player.ID] += action.Amount
		h.State.CurrentBet = action.Amount
		h.State.LastRaisePosition = h.State.CurrentTurn

		log.Printf("[INFO] Player %s bets %d", player.Player.ID, action.Amount)

		// If player is all-in, log it
		if player.Balance == 0 {
			log.Printf("[INFO] Player %s is all-in", player.Player.ID)
		}

	case HoldemActionRaise:
		if h.State.CurrentBet == 0 {
			return errors.New("cannot raise, must bet")
		}

		// Calculate minimum raise
		minRaise := h.State.CurrentBet*2 - playerBet

		if action.Amount < minRaise {
			return errors.New("raise must be at least double the current bet")
		}

		if action.Amount > player.Balance {
			return errors.New("insufficient balance to raise")
		}

		// Handle the raise
		toRaise := action.Amount - playerBet
		player.Balance -= toRaise
		h.State.Pot += toRaise
		h.State.PlayerBets[player.Player.ID] += toRaise
		h.State.CurrentBet = action.Amount
		h.State.LastRaisePosition = h.State.CurrentTurn

		log.Printf("[INFO] Player %s raises to %d", player.Player.ID, action.Amount)

		// If player is all-in, log it
		if player.Balance == 0 {
			log.Printf("[INFO] Player %s is all-in", player.Player.ID)
		}

	case HoldemActionAllIn:
		if player.Balance == 0 {
			return errors.New("player already all-in")
		}

		// All-in amount
		allInAmount := player.Balance + playerBet

		h.State.Pot += player.Balance
		h.State.PlayerBets[player.Player.ID] += player.Balance

		// If the all-in amount is a raise, update the current bet and last raiser
		if allInAmount > h.State.CurrentBet {
			h.State.CurrentBet = allInAmount
			h.State.LastRaisePosition = h.State.CurrentTurn
		}

		player.Balance = 0
		log.Printf("[INFO] Player %s goes all-in with %d", player.Player.ID, allInAmount)

	default:
		return errors.New("invalid action")
	}

	h.LogGameState(fmt.Sprintf("AFTER %s ACTION BY %s", action.Action, player.Player.ID))
	h.State.CurrentTurn = h.NextActivePlayerAfter(h.State.CurrentTurn)
	if h.CheckRoundComplete() {
		h.State.RoundComplete = true
	}

	return nil
}

func (h *Holdem) DealCards() error {
	for _, player := range h.game.Players {
		card1, err := h.deck.Draw()
		if err != nil {
			return err
		}
		card2, err := h.deck.Draw()
		if err != nil {
			return err
		}

		h.game.Mu.Lock()
		h.State.PlayerHands[player.Player.ID] = append(h.State.PlayerHands[player.Player.ID], card1, card2)
		h.game.Mu.Unlock()
		log.Printf("[INFO] Dealt cards to player %s: %v", player.Player.ID, h.State.PlayerHands[player.Player.ID])
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

	log.Printf("[INFO] Dealt community cards: %v", communityCards)

	return nil
}

func (h *Holdem) PlayRound() {
	log.Printf("[INFO] Starting new round")
	h.HandlePlayers()
	log.Printf("[INFO] Players after handling: %v", &h.game.Players)
	h.deck = models.NewDeck()

	h.StartPreFlopRound()
	if h.PlayersNotFoldedCount() <= 1 {
		log.Printf("[INFO] Only one player remains - hand complete")
		h.EvaluateHands()

		if h.CanGameContinue() {
			h.PlayRound()
		} else {
			h.End()
		}
		return
	}

	communityCards := h.State.CommunityCards
	h.StartFlopRound(communityCards)
	if h.PlayersNotFoldedCount() <= 1 {
		log.Printf("[INFO] Only one player remains - hand complete")
		h.EvaluateHands()

		if h.CanGameContinue() {
			h.PlayRound()
		} else {
			h.End()
		}
		return
	}

	h.StartTurnRound(communityCards)
	if h.PlayersNotFoldedCount() <= 1 {
		log.Printf("[INFO] Only one player remains - hand complete")
		h.EvaluateHands()

		if h.CanGameContinue() {
			h.PlayRound()
		} else {
			h.End()
		}
		return
	}

	h.StartRiverRound(communityCards)
	h.State.CurrentRound = Showdown
	h.LogGameState("SHOWDOWN")
	h.EvaluateHands()
	h.LogGameState("HAND COMPLETE")

	if h.CanGameContinue() {
		h.PlayRound()
	} else {
		h.End()
	}
}

func (h *Holdem) StartPreFlopRound() {
	h.State.Pot = 0
	h.State.CurrentBet = 0
	h.State.CurrentRound = PreFlop
	h.State.RoundComplete = false
	h.State.LastRaisePosition = -1

	h.RotateDealerButton()
	if !h.SetBlindPositions() {
		log.Printf("[ERROR] Not enough players to set blind positions")
		h.End()
		return
	}

	if err := h.DealCards(); err != nil {
		log.Printf("[ERROR] Failed to deal cards: %v", err)
		h.End()
		return
	}

	h.PostBlinds()
	if len(h.game.Players) > 2 {
		h.State.CurrentTurn = (h.State.BigBlindIndex + 1) % len(h.game.Players)
	} else { // In heads-up play, small blind acts first pre-flop
		h.State.CurrentTurn = h.State.SmallBlindIndex
	}

	h.LogGameState("HAND STARTED - PRE-FLOP BETTING BEGINS")
	log.Printf("[INFO] Starting pre-flop betting")
	if err := h.BettingRound(); err != nil {
		log.Printf("[ERROR] Betting round error: %v", err)
		h.End()
		return
	}
	h.LogGameState("PRE-FLOP BETTING COMPLETE")
}

func (h *Holdem) StartFlopRound(communityCards []models.Card) {
	h.State.CurrentRound = Flop
	h.State.RoundComplete = false
	h.State.CommunityCards = communityCards[:3]
	h.State.CurrentBet = 0
	// In post-flop betting rounds, action starts with first active player after dealer
	h.State.CurrentTurn = h.NextActivePlayerAfter(h.State.DealerIndex)
	h.State.LastRaisePosition = -1

	h.LogGameState("FLOP DEALT - FLOP BETTING BEGINS")
	if err := h.BettingRound(); err != nil {
		log.Printf("[ERROR] Betting round error: %v", err)
		h.End()
		return
	}
	h.LogGameState("FLOP BETTING COMPLETE")
}

func (h *Holdem) StartTurnRound(communityCards []models.Card) {
	h.State.CurrentRound = Turn
	h.State.RoundComplete = false
	h.State.CommunityCards = communityCards[:4]
	h.State.CurrentBet = 0
	h.State.CurrentTurn = h.NextActivePlayerAfter(h.State.DealerIndex)
	h.State.LastRaisePosition = -1

	h.LogGameState("TURN DEALT - TURN BETTING BEGINS")
	if err := h.BettingRound(); err != nil {
		log.Printf("[ERROR] Betting round error: %v", err)
		h.End()
		return
	}
	h.LogGameState("TURN BETTING COMPLETE")
}

func (h *Holdem) StartRiverRound(communityCards []models.Card) {
	h.State.CurrentRound = River
	h.State.RoundComplete = false
	h.State.CommunityCards = communityCards[:5]
	h.State.CurrentBet = 0
	h.State.CurrentTurn = h.NextActivePlayerAfter(h.State.DealerIndex)
	h.State.LastRaisePosition = -1

	h.LogGameState("RIVER DEALT - RIVER BETTING BEGINS")
	if err := h.BettingRound(); err != nil {
		log.Printf("[ERROR] Betting round error: %v", err)
		h.End()
		return
	}
	h.LogGameState("RIVER BETTING COMPLETE")
}

func (h *Holdem) PostBlinds() {
	smallBlindPlayer := h.game.Players[h.State.SmallBlindIndex]
	smallBlindAmount := min(h.State.SmallBlindAmount, smallBlindPlayer.Balance)

	smallBlindPlayer.Balance -= smallBlindAmount
	h.State.Pot += smallBlindAmount
	h.State.PlayerBets[smallBlindPlayer.Player.ID] = smallBlindAmount

	bigBlindPlayer := h.game.Players[h.State.BigBlindIndex]
	bigBlindAmount := min(h.State.BigBlindAmount, bigBlindPlayer.Balance)

	bigBlindPlayer.Balance -= bigBlindAmount
	h.State.Pot += bigBlindAmount
	h.State.CurrentBet = bigBlindAmount
	h.State.PlayerBets[bigBlindPlayer.Player.ID] = bigBlindAmount

	log.Printf("[BLINDS] Player %s posts small blind %d", smallBlindPlayer.Player.ID, smallBlindAmount)
	log.Printf("[BLINDS] Player %s posts big blind: %d", bigBlindPlayer.Player.ID, bigBlindAmount)
}

func (h *Holdem) NextActivePlayerAfter(pos int) int {
	players := h.GetPlayersInRound()
	numPlayers := len(players)
	nextPlayerIndex := (pos + 1) % numPlayers

	return nextPlayerIndex
}

func (h *Holdem) BettingRound() error {
	log.Printf("[INFO] Starting betting round for %v", h.State.CurrentRound)
	h.State.PlayerBets = make(map[string]int)
	for _, player := range h.game.Players {
		h.State.PlayerBets[player.Player.ID] = 0
	}

	h.State.RoundComplete = false
	for !h.State.RoundComplete {
		log.Printf("[DEBUG] Round state - CurrentTurn: %d, RoundComplete: %v", h.State.CurrentTurn, h.State.RoundComplete)
		if h.CheckRoundComplete() {
			h.State.RoundComplete = true
			log.Printf("[DEBUG] Round completed naturally")
			break
		}

		if h.State.CurrentTurn >= len(h.game.Players) {
			h.State.CurrentTurn = 0
		}

		player := h.game.Players[h.State.CurrentTurn]
		if h.State.PlayerHands[player.Player.ID] == nil || player.Balance == 0 { // Skip players who have folded or are all-in
			h.State.CurrentTurn = (h.State.CurrentTurn + 1) % len(h.game.Players)
			if h.CheckRoundComplete() {
				h.State.RoundComplete = true
			}

			continue
		}

		log.Printf("[INFO] Waiting for action from player %s", player.Player.ID)

		playerBet := h.State.PlayerBets[player.Player.ID]
		toCall := h.State.CurrentBet - playerBet
		log.Printf("[ACTION] Player %s to act | Current bet: $%d | Player bet: $%d | To call: $%d | Balance: $%d", player.Player.ID, h.State.CurrentBet, playerBet, toCall, player.Balance)

		timer := time.NewTimer(5 * time.Second)
		actionReceived := make(chan bool, 1)
		go func() {
			select {
			case action := <-h.actionsChannel:
				if action.PlayerID == player.Player.ID {
					if err := h.ProcessAction(action.Data); err != nil {
						log.Printf("[ERROR] Failed to process action: %v", err)

						fold := HoldemAction{
							PlayerID: player.Player.ID,
							Action:   HoldemActionFold,
						}
						foldData, _ := json.Marshal(fold)
						h.ProcessAction(foldData)
					}
					actionReceived <- true
				}
			case <-timer.C:
				log.Printf("[INFO] Player %s timed out", player.Player.ID)

				// Auto-check if possible, otherwise fold
				playerBet := h.State.PlayerBets[player.Player.ID]
				toCall := h.State.CurrentBet - playerBet

				var autoAction HoldemAction
				if toCall == 0 {
					autoAction = HoldemAction{
						PlayerID: player.Player.ID,
						Action:   HoldemActionCheck,
					}
				} else if toCall <= player.Balance {
					autoAction = HoldemAction{
						PlayerID: player.Player.ID,
						Action:   HoldemActionCall,
					}
				} else {
					autoAction = HoldemAction{
						PlayerID: player.Player.ID,
						Action:   HoldemActionFold,
					}
				}

				autoData, _ := json.Marshal(autoAction)
				h.ProcessAction(autoData)
				actionReceived <- true
			}
		}()

		<-actionReceived
		timer.Stop()
	}

	log.Printf("[INFO] Betting round complete for %v", h.State.CurrentRound)
	return nil
}

func (h *Holdem) DealPlayerCards() error {
	h.deck.Shuffle()

	// Clear any existing hands
	for _, player := range h.game.Players {
		h.State.PlayerHands[player.Player.ID] = make([]models.Card, 0, 2)
	}

	// Deal cards in the correct order (one card at a time, starting with player after dealer)
	startPos := (h.State.DealerIndex + 1) % len(h.game.Players)

	// First round of cards
	for i := 0; i < len(h.game.Players); i++ {
		playerPos := (startPos + i) % len(h.game.Players)
		player := h.game.Players[playerPos]

		if player.Status == models.GamePlayerStatusActive {
			card, err := h.deck.Draw()
			if err != nil {
				return err
			}
			h.State.PlayerHands[player.Player.ID] = append(h.State.PlayerHands[player.Player.ID], card)
		}
	}

	// Second round of cards
	for i := 0; i < len(h.game.Players); i++ {
		playerPos := (startPos + i) % len(h.game.Players)
		player := h.game.Players[playerPos]

		if player.Status == models.GamePlayerStatusActive {
			card, err := h.deck.Draw()
			if err != nil {
				return err
			}
			h.State.PlayerHands[player.Player.ID] = append(h.State.PlayerHands[player.Player.ID], card)
		}
	}

	// Log player hands (for debugging)
	for _, player := range h.game.Players {
		if player.Status == models.GamePlayerStatusActive {
			log.Printf("[INFO] Player %s received cards: %v", player.Player.ID, h.State.PlayerHands[player.Player.ID])
		}
	}

	return nil
}

func (h *Holdem) DealCommunityCards() error {
	// Clear any existing community cards
	h.State.CommunityCards = []models.Card{}

	// Burn a card before the flop (standard poker rule)
	_, err := h.deck.Draw()
	if err != nil {
		return err
	}

	// Deal the flop (3 cards)
	for i := 0; i < 3; i++ {
		card, err := h.deck.Draw()
		if err != nil {
			return err
		}
		h.State.CommunityCards = append(h.State.CommunityCards, card)
	}

	// Burn a card before the turn
	_, err = h.deck.Draw()
	if err != nil {
		return err
	}

	// Deal the turn (1 card)
	card, err := h.deck.Draw()
	if err != nil {
		return err
	}
	h.State.CommunityCards = append(h.State.CommunityCards, card)

	// Burn a card before the river
	_, err = h.deck.Draw()
	if err != nil {
		return err
	}

	// Deal the river (1 card)
	card, err = h.deck.Draw()
	if err != nil {
		return err
	}
	h.State.CommunityCards = append(h.State.CommunityCards, card)

	// Log the community cards (for debugging)
	log.Printf("[INFO] Community cards (hidden until revealed): %v", h.State.CommunityCards)

	return nil
}

func (h *Holdem) HandlePlayers() {
	h.game.Mu.Lock()
	defer h.game.Mu.Unlock()

	filteredPlayers := []*models.GamePlayer{}
	for _, player := range h.game.Players {
		if player.Status == models.GamePlayerStatusInactive || player.Balance < h.game.MinBet {
			continue
		}

		if player.Status == models.GamePlayerStatusWaiting {
			player.Status = models.GamePlayerStatusActive
		}

		h.State.PlayerHands[player.Player.ID] = make([]models.Card, 0)
		filteredPlayers = append(filteredPlayers, player)
	}

	sort.Slice(filteredPlayers, func(i, j int) bool {
		return filteredPlayers[i].Position < filteredPlayers[j].Position
	})

	h.game.Players = filteredPlayers
}

func (h *Holdem) CheckRoundComplete() bool {
	if h.PlayersNotFoldedCount() <= 1 {
		return true
	}

	activePlayers := 0
	for _, player := range h.game.Players {
		if h.State.PlayerHands[player.Player.ID] != nil && player.Balance > 0 {
			activePlayers++
		}
	}

	if activePlayers == 0 {
		return true
	}

	// Round has completed when:
	// 1. Everyone has had a chance to act at least once, AND
	// 2. Everyone has had a chance to respond to the last raise

	// For pre-flop, special handling because of blinds
	if h.State.CurrentRound == PreFlop {
		// If no one has raised beyond the big blind
		if h.State.LastRaisePosition == -1 {
			// The round is complete when we reach the big blind position
			return h.State.CurrentTurn == h.State.BigBlindIndex
		}
	} else {
		// For post-flop, if no bets made, round is complete when it goes full circle
		if h.State.LastRaisePosition == -1 && h.State.CurrentBet == 0 {
			startPos := (h.State.DealerIndex + 1) % len(h.game.Players)
			return h.State.CurrentTurn == startPos
		}
	}

	// If there was a raise, the round completes when action gets back to the last raiser
	if h.State.LastRaisePosition != -1 {
		return h.State.CurrentTurn == h.State.LastRaisePosition
	}

	return false
}

func (h *Holdem) CanGameContinue() bool {
	// Check if there are at least 2 players with chips
	playersWithChips := 0
	for _, player := range h.game.Players {
		if player.Balance >= h.game.MinBet && player.Status == models.GamePlayerStatusActive {
			playersWithChips++
		}
	}

	return playersWithChips >= 2
}

func (h *Holdem) Start() error {
	if !h.CanStart() {
		return errors.New("not enough players to start")
	}

	h.game.Status = models.GameStatusStarted
	log.Printf("[INFO] Starting holdem game")

	for _, player := range h.game.Players { // Initialize all waiting players as active with starting chips
		if player.Status == models.GamePlayerStatusWaiting {
			player.Status = models.GamePlayerStatusActive
		}
	}

	h.LogGameState("GAME STARTING")
	go h.PlayRound()

	return nil
}

func (h *Holdem) End() error {
	h.game.Status = models.GameStatusEnd
	log.Printf("[INFO] Ending holdem game")

	// Log final game state
	h.LogGameState("GAME ENDED")

	// Notify players of game end
	// (This would be handled by your notification system)

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

func (h *Holdem) PlayersNotFoldedCount() int {
	count := 0
	for _, player := range h.game.Players {
		if h.State.PlayerHands[player.Player.ID] != nil {
			count++
			continue
		}

		if len(h.State.PlayerHands[player.Player.ID]) > 0 {
			count++
		}
	}

	return count
}

func (h *Holdem) OnPlayerJoin(player *models.GamePlayer) error {
	player.Status = models.GamePlayerStatusWaiting
	log.Printf("[INFO] Player %s joined the game", player.Player.ID)

	if h.game.Status == models.GameStatusWaiting && h.CanStart() {
		err := h.Start()
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *Holdem) OnPlayerLeave(player *models.GamePlayer) error {
	player.Status = models.GamePlayerStatusInactive
	log.Printf("[INFO] Player %s left the game", player.Player.ID)

	return nil
}

// // Helper function to handle all-in situations and side pots
// func (h *Holdem) HandleAllIns() {
// 	// Reset side pots
// 	h.State.SidePots = []SidePot{}

// 	// Get all active players (not folded)
// 	activePlayers := []*models.GamePlayer{}
// 	for _, player := range h.game.Players {
// 		if player.Hand != nil {
// 			activePlayers = append(activePlayers, player)
// 		}
// 	}

// 	// If only one player or no players, no need for side pots
// 	if len(activePlayers) <= 1 {
// 		return
// 	}

// 	// Sort players by their chip contribution to the pot (all-in players first)
// 	// This would require tracking how much each player has put in the pot
// 	// For simplicity, this implementation is left as an exercise
// }

// Rotate dealer button
func (h *Holdem) RotateDealerButton() {
	activePlayers := h.GetPlayersInRound()
	if len(activePlayers) == 0 {
		return
	}

	currentDealerIndex := h.State.DealerIndex
	if currentDealerIndex == -1 || h.game.Status == models.GameStatusStarting {
		h.State.DealerIndex = 0
	} else {
		nextDealerIndex := (currentDealerIndex + 1) % len(activePlayers)
		h.State.DealerIndex = nextDealerIndex
	}
}

// Set blind positions based on active players
func (h *Holdem) SetBlindPositions() bool {
	activePlayers := h.GetPlayersInRound()
	if len(activePlayers) < 2 {
		return false
	}

	if len(activePlayers) == 2 { // In heads-up (2 players), dealer is small blind
		h.State.SmallBlindIndex = h.State.DealerIndex
		h.State.BigBlindIndex = (h.State.DealerIndex + 1) % 2
	} else { // With more players, small blind is after dealer
		h.State.SmallBlindIndex = (h.State.DealerIndex + 1) % len(activePlayers)
		h.State.BigBlindIndex = (h.State.DealerIndex + 2) % len(activePlayers)
	}

	return true
}

// EvaluateHands evaluates all player hands and determines the winner(s)
func (h *Holdem) EvaluateHands() error {
	log.Println("[INFO] Evaluating hands")
	activePlayers := []*models.GamePlayer{}
	for _, player := range h.game.Players {
		if h.State.PlayerHands[player.Player.ID] != nil && len(h.State.PlayerHands[player.Player.ID]) > 0 {
			activePlayers = append(activePlayers, player)
		}
	}

	if len(activePlayers) == 0 {
		log.Println("[ERROR] No active players to evaluate hands")
		return nil
	}

	if len(activePlayers) == 1 {
		winner := activePlayers[0]
		winner.Balance += h.State.Pot
		log.Printf("[INFO] Player %s wins %d (uncontested)", winner.Player.ID, h.State.Pot)
		h.State.Pot = 0
		return nil
	}

	results := []HandResult{}
	for _, player := range activePlayers {
		allCards := append([]models.Card{}, h.State.PlayerHands[player.Player.ID]...)
		allCards = append(allCards, h.State.CommunityCards...)

		// Get best 5-card hand
		handRank, highCards := evaluateBestHand(allCards)

		results = append(results, HandResult{
			Rank:      handRank,
			HighCards: highCards,
			PlayerID:  player.Player.ID,
		})

		log.Printf("[INFO] Player %s has %s", player.Player.ID, handRank)
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Rank != results[j].Rank {
			return results[i].Rank > results[j].Rank
		}

		for k := 0; k < len(results[i].HighCards) && k < len(results[j].HighCards); k++ {
			if results[i].HighCards[k] != results[j].HighCards[k] {
				return results[i].HighCards[k] > results[j].HighCards[k]
			}
		}

		// Completely equal hands
		return false
	})

	winners := []HandResult{results[0]}
	for i := 1; i < len(results); i++ {
		if compareHands(results[i], results[0]) == 0 {
			winners = append(winners, results[i])
		} else {
			break
		}
	}

	// Distribute pot among winners
	winAmount := h.State.Pot / len(winners)
	remaining := h.State.Pot % len(winners)

	for _, winner := range winners {
		for _, player := range h.game.Players {
			if player.Player.ID == winner.PlayerID {
				player.Balance += winAmount
				// Give any remainder to the first winner (arbitrary but consistent)
				if remaining > 0 && player.Player.ID == winners[0].PlayerID {
					player.Balance += remaining
					remaining = 0
				}
				break
			}
		}
	}

	if len(winners) == 1 {
		log.Printf("[INFO] Player %s wins %d with %s", winners[0].PlayerID, h.State.Pot, winners[0].Rank)
	} else {
		winnerIDs := ""
		for i, winner := range winners {
			if i > 0 {
				winnerIDs += ", "
			}
			winnerIDs += winner.PlayerID
		}
		log.Printf("[INFO] Players %s split the pot (%d each) with %s", winnerIDs, winAmount, winners[0].Rank)
	}

	h.State.Pot = 0

	return nil
}

// Returns: 1 if hand1 > hand2, 0 if equal, -1 if hand1 < hand2
func compareHands(hand1, hand2 HandResult) int {
	// Compare hand ranks
	if hand1.Rank > hand2.Rank {
		return 1
	}
	if hand1.Rank < hand2.Rank {
		return -1
	}

	// Same rank, compare high cards
	for i := 0; i < len(hand1.HighCards) && i < len(hand2.HighCards); i++ {
		if hand1.HighCards[i] > hand2.HighCards[i] {
			return 1
		}
		if hand1.HighCards[i] < hand2.HighCards[i] {
			return -1
		}
	}

	// Hands are equal
	return 0
}

// Evaluate the best 5-card hand from the given cards
func evaluateBestHand(cards []models.Card) (HandRank, []int) {
	if rank, highCards := checkStraightFlush(cards); rank != HighCard {
		return rank, highCards
	}

	if highCards := checkFourOfAKind(cards); highCards != nil {
		return FourOfAKind, highCards
	}

	if highCards := checkFullHouse(cards); highCards != nil {
		return FullHouse, highCards
	}

	if highCards := checkFlush(cards); highCards != nil {
		return Flush, highCards
	}

	if highCards := checkStraight(cards); highCards != nil {
		return Straight, highCards
	}

	if highCards := checkThreeOfAKind(cards); highCards != nil {
		return ThreeOfAKind, highCards
	}

	if highCards := checkTwoPair(cards); highCards != nil {
		return TwoPair, highCards
	}

	if highCards := checkOnePair(cards); highCards != nil {
		return OnePair, highCards
	}

	return HighCard, getHighCards(cards, 5)
}

func checkStraightFlush(cards []models.Card) (HandRank, []int) {
	cardsBySuit := make(map[string][]models.Card)
	for _, card := range cards {
		cardsBySuit[card.Suit] = append(cardsBySuit[card.Suit], card)
	}

	for _, suitCards := range cardsBySuit {
		if len(suitCards) >= 5 {
			if highCard := checkStraight(suitCards); highCard != nil {
				if highCard[0] == 14 {
					return RoyalFlush, highCard
				}
				return StraightFlush, highCard
			}
		}
	}

	return HighCard, nil
}

func checkFourOfAKind(cards []models.Card) []int {
	valueCounts := make(map[int]int)
	for _, card := range cards {
		valueCounts[card.Value]++
	}

	var fourOfAKind int
	for value, count := range valueCounts {
		if count >= 4 {
			fourOfAKind = value
			break
		}
	}

	if fourOfAKind > 0 {
		kicker := 0
		for value := range valueCounts {
			if value != fourOfAKind && value > kicker {
				kicker = value
			}
		}

		return []int{fourOfAKind, kicker}
	}

	return nil
}

func checkFullHouse(cards []models.Card) []int {
	valueCounts := make(map[int]int)
	for _, card := range cards {
		valueCounts[card.Value]++
	}

	var threeKind, pair int
	for value, count := range valueCounts {
		if count >= 3 && value > threeKind {
			threeKind = value
		}
	}

	for value, count := range valueCounts {
		if value != threeKind && count >= 2 && value > pair {
			pair = value
		}
	}

	if threeKind > 0 && pair > 0 {
		return []int{threeKind, pair}
	}

	return nil
}

func checkFlush(cards []models.Card) []int {
	cardsBySuit := make(map[string][]models.Card)
	for _, card := range cards {
		cardsBySuit[card.Suit] = append(cardsBySuit[card.Suit], card)
	}

	for _, suitCards := range cardsBySuit {
		if len(suitCards) >= 5 {
			return getHighCards(suitCards, 5)
		}
	}

	return nil
}

func checkStraight(cards []models.Card) []int {
	valueSet := make(map[int]bool)
	for _, card := range cards {
		valueSet[card.Value] = true
	}

	if valueSet[14] && valueSet[5] && valueSet[4] && valueSet[3] && valueSet[2] {
		return []int{5} // 5-high straight
	}

	values := make([]int, 0, len(valueSet))
	for value := range valueSet {
		values = append(values, value)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(values)))

	for i := 0; i <= len(values)-5; i++ {
		if values[i]-values[i+4] == 4 {
			return []int{values[i]} // Return highest card
		}
	}

	return nil
}

func checkThreeOfAKind(cards []models.Card) []int {
	valueCounts := make(map[int]int)
	for _, card := range cards {
		valueCounts[card.Value]++
	}

	var threeKind int
	for value, count := range valueCounts {
		if count >= 3 && value > threeKind {
			threeKind = value
		}
	}

	if threeKind > 0 {
		kickers := []int{}
		for value := range valueCounts {
			if value != threeKind {
				kickers = append(kickers, value)
			}
		}

		sort.Sort(sort.Reverse(sort.IntSlice(kickers)))
		if len(kickers) > 2 {
			kickers = kickers[:2]
		}

		return append([]int{threeKind}, kickers...)
	}

	return nil
}

func checkTwoPair(cards []models.Card) []int {
	valueCounts := make(map[int]int)
	for _, card := range cards {
		valueCounts[card.Value]++
	}

	pairs := []int{}
	for value, count := range valueCounts {
		if count >= 2 {
			pairs = append(pairs, value)
		}
	}

	if len(pairs) >= 2 {
		sort.Sort(sort.Reverse(sort.IntSlice(pairs)))

		// Take highest two pairs
		highPairs := pairs
		if len(highPairs) > 2 {
			highPairs = pairs[:2]
		}

		kicker := 0
		for value := range valueCounts {
			if value != highPairs[0] && value != highPairs[1] && value > kicker {
				kicker = value
			}
		}

		return append(highPairs, kicker)
	}

	return nil
}

func checkOnePair(cards []models.Card) []int {
	valueCounts := make(map[int]int)
	for _, card := range cards {
		valueCounts[card.Value]++
	}

	var pair int
	for value, count := range valueCounts {
		if count >= 2 && value > pair {
			pair = value
		}
	}

	if pair > 0 {
		kickers := []int{}
		for value := range valueCounts {
			if value != pair {
				kickers = append(kickers, value)
			}
		}

		sort.Sort(sort.Reverse(sort.IntSlice(kickers)))
		if len(kickers) > 3 {
			kickers = kickers[:3]
		}

		return append([]int{pair}, kickers...)
	}

	return nil
}

func getHighCards(cards []models.Card, count int) []int {
	values := []int{}
	for _, card := range cards {
		values = append(values, card.Value)
	}

	sort.Sort(sort.Reverse(sort.IntSlice(values)))

	// Remove duplicates
	unique := []int{}
	seen := make(map[int]bool)

	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			unique = append(unique, value)
		}
	}

	// Return top n values
	if len(unique) > count {
		return unique[:count]
	}
	return unique
}

func (h *Holdem) GetPlayersInRound() []*models.GamePlayer {
	activePlayers := []*models.GamePlayer{}

	for _, player := range h.game.Players {
		if player.Status == models.GamePlayerStatusActive || player.Status == models.GamePlayerStatusInactive {
			activePlayers = append(activePlayers, player)
		}
	}

	return activePlayers
}

func (h *Holdem) GetActivePlayers() []*models.GamePlayer {
	activePlayers := []*models.GamePlayer{}

	for _, player := range h.game.Players {
		if player.Status == models.GamePlayerStatusActive {
			activePlayers = append(activePlayers, player)
		}
	}

	return activePlayers
}

// Return current game state (for sending to clients)
func (h *Holdem) GetGameState() interface{} {
	// Create a sanitized version of the state to send to players
	// This will hide information that players shouldn't see
	type PlayerView struct {
		ID            string
		Name          string
		Balance       int
		HasCards      bool
		Folded        bool
		IsDealer      bool
		IsSmallBlind  bool
		IsBigBlind    bool
		IsCurrentTurn bool
	}

	type GameView struct {
		Players        []PlayerView
		CommunityCards []models.Card
		Pot            int
		CurrentBet     int
		CurrentRound   string
	}

	// Convert round to string
	roundName := "preflop"
	switch h.State.CurrentRound {
	case Flop:
		roundName = "flop"
	case Turn:
		roundName = "turn"
	case River:
		roundName = "river"
	case Showdown:
		roundName = "showdown"
	}

	// Only show community cards appropriate for the current round
	visibleCards := []models.Card{}
	switch h.State.CurrentRound {
	case Flop:
		visibleCards = h.State.CommunityCards[:3]
	case Turn:
		visibleCards = h.State.CommunityCards[:4]
	case River, Showdown:
		visibleCards = h.State.CommunityCards[:5]
	}

	// Build player views
	playerViews := []PlayerView{}
	for i, player := range h.game.Players {
		playerViews = append(playerViews, PlayerView{
			ID:            player.Player.ID,
			Name:          player.Player.Username,
			Balance:       player.Balance,
			HasCards:      len(h.State.PlayerHands[player.Player.ID]) > 0,
			Folded:        h.State.PlayerHands[player.Player.ID] == nil,
			IsDealer:      i == h.State.DealerIndex,
			IsSmallBlind:  i == h.State.SmallBlindIndex,
			IsBigBlind:    i == h.State.BigBlindIndex,
			IsCurrentTurn: i == h.State.CurrentTurn,
		})
	}

	return GameView{
		Players:        playerViews,
		CommunityCards: visibleCards,
		Pot:            h.State.Pot,
		CurrentBet:     h.State.CurrentBet,
		CurrentRound:   roundName,
	}
}

// Get player-specific game state (includes their cards)
func (h *Holdem) GetPlayerState(playerID string) interface{} {
	baseState := h.GetGameState().(map[string]interface{})

	// Find the player's hand
	var playerHand []models.Card
	for _, player := range h.game.Players {
		if player.Player.ID == playerID {
			playerHand = h.State.PlayerHands[player.Player.ID]
			break
		}
	}

	// Add player's hand to the state
	baseState["hand"] = playerHand

	return baseState
}

func (h *Holdem) LogGameState(message string) {
	separator := "===================================================="
	log.Printf("\n%s\n[GAME STATE] %s\n%s", separator, message, separator)

	roundNames := map[HoldemRound]string{
		PreFlop:  "PRE-FLOP",
		Flop:     "FLOP",
		Turn:     "TURN",
		River:    "RIVER",
		Showdown: "SHOWDOWN",
	}
	log.Printf("[ROUND] Current Round: %s", roundNames[h.State.CurrentRound])
	log.Printf("[POT] Total Pot: $%d | Current Bet: $%d", h.State.Pot, h.State.CurrentBet)

	if h.State.CurrentRound >= Flop {
		communityCards := ""
		switch h.State.CurrentRound {
		case Flop:
			communityCards = fmt.Sprintf("%v", h.State.CommunityCards[:3])
		case Turn:
			communityCards = fmt.Sprintf("%v", h.State.CommunityCards[:4])
		case River, Showdown:
			communityCards = fmt.Sprintf("%v", h.State.CommunityCards[:5])
		}
		log.Printf("[COMMUNITY] Cards: %s", communityCards)
	}

	log.Printf("[PLAYERS] Status:")
	for i, player := range h.game.Players {
		status := "Active"
		if h.State.PlayerHands[player.Player.ID] == nil {
			status = "Folded"
		} else if player.Balance == 0 {
			status = "All-In"
		}

		position := ""
		if i == h.State.DealerIndex {
			position += "Dealer "
		}
		if i == h.State.SmallBlindIndex {
			position += "SB "
		}
		if i == h.State.BigBlindIndex {
			position += "BB "
		}
		if i == h.State.CurrentTurn {
			position += "Acting "
		}

		playerBet := h.State.PlayerBets[player.Player.ID]

		log.Printf("  - Player %s (%s):", player.Player.ID, position)
		log.Printf("	- Balance: $%d | Bet: $%d | Status: %s ", player.Balance, playerBet, status)
		log.Printf("	- Cards: %v", func() string {
			if h.State.PlayerHands[player.Player.ID] == nil {
				return "Folded"
			}
			return fmt.Sprintf("%v", h.State.PlayerHands[player.Player.ID])
		}())
	}

	activePlayers := h.PlayersNotFoldedCount()
	log.Printf("[ACTIVE] Players still in hand: %d", activePlayers)

	log.Printf("%s\n", separator)
}

func Where[T any](players []T, condition func(player T) bool) []T {
	filteredPlayers := []T{}
	for _, player := range players {
		if condition(player) {
			filteredPlayers = append(filteredPlayers, player)
		}
	}
	return filteredPlayers
}

type HandRank int

const (
	HighCard HandRank = iota
	OnePair
	TwoPair
	ThreeOfAKind
	Straight
	Flush
	FullHouse
	FourOfAKind
	StraightFlush
	RoyalFlush
)

type HandResult struct {
	Rank      HandRank
	HighCards []int // Cards used to break ties, in descending order of importance
	PlayerID  string
}
