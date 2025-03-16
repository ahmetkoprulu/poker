package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"slices"
	"sort"
	"sync"
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

type HoldemMessageType string

const (
	HoldemMessageGameStart    HoldemMessageType = "game_start"
	HoldemMessageGameEnd      HoldemMessageType = "game_end"
	HoldemMessageRoundStart   HoldemMessageType = "round_start"
	HoldemMessageRoundEnd     HoldemMessageType = "round_end"
	HoldemMessagePlayerTurn   HoldemMessageType = "player_turn"
	HoldemMessagePlayerAction HoldemMessageType = "player_action"
	HoldemMessageShowdown     HoldemMessageType = "showdown"
	HoldemMessageWinner       HoldemMessageType = "winner"
)

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

type HoldemMessage struct {
	Type      HoldemMessageType `json:"type"`
	Data      interface{}       `json:"data"`
	Timestamp int64             `json:"timestamp"`
}

type Holdem struct {
	State          HoldemState
	deck           *models.Deck
	game           *Game
	actionsChannel chan GameAction
	messageChannel chan GameMessage
	doneChannel    chan bool

	Mu sync.RWMutex
}

type HoldemAction struct {
	PlayerID string           `json:"playerId"`
	Action   HoldemActionType `json:"action"`
	Amount   int              `json:"amount"`
}

type HoldemState struct {
	Pot              int
	CurrentRound     HoldemRound
	CurrentBet       int
	BigBlindAmount   int
	SmallBlindAmount int
	RoundComplete    bool

	DealerSeat     *TableSeat
	SmallBlindSeat *TableSeat
	BigBlindSeat   *TableSeat
	CurrentSeat    *TableSeat
	LastRaiserSeat *TableSeat
	Seats          map[int]*TableSeat

	PlayerBets       map[string]int
	PlayerLastAction map[string]HoldemActionType
	CommunityCards   []models.Card

	// Side pots for all-in situations
	// SidePots []SidePot
}

type TableSeat struct {
	Position int
	Player   *GamePlayer
	Next     *TableSeat
	Prev     *TableSeat
	Hand     []models.Card
}

type HandResult struct {
	Rank      HandRank
	HighCards []int
	PlayerID  string
}

type HoldemRound int

const (
	PreFlop HoldemRound = iota
	Flop
	Turn
	River
	Showdown
)

type HoldemPlayerActionMessage struct {
	PlayerID  string           `json:"player_id"`
	Action    HoldemActionType `json:"action"`
	Amount    int              `json:"amount"`
	GameState interface{}      `json:"game_state"`
}

type HoldemWinnerMessage struct {
	WinnerID string `json:"winner_id"`
	Amount   int    `json:"amount"`
	Reason   string `json:"reason"`
}

type HoldemShowdownMessage struct {
	Winners   []HandResult `json:"winners"`
	Pot       int          `json:"pot"`
	GameState interface{}  `json:"game_state"`
}

type HoldemPlayerTurnMessage struct {
	GameState interface{} `json:"game_state"`
	Timeout   int         `json:"timeout"`
}

func NewHoldem(game *Game) *Holdem {
	return &Holdem{
		State: HoldemState{
			SmallBlindAmount: 0,
			BigBlindAmount:   0,
			CurrentRound:     PreFlop,
			Pot:              0,
			CurrentBet:       0,
			RoundComplete:    false,

			DealerSeat:     nil,
			CurrentSeat:    nil,
			SmallBlindSeat: nil,
			BigBlindSeat:   nil,
			LastRaiserSeat: nil,

			Seats:          make(map[int]*TableSeat),
			PlayerBets:     make(map[string]int),
			CommunityCards: make([]models.Card, 0),
		},
		actionsChannel: game.ActionChan,
		messageChannel: game.MessageChan,
		doneChannel:    make(chan bool),
		deck:           models.NewDeck(),
		game:           game,
		Mu:             sync.RWMutex{},
	}
}

func (h *Holdem) RefreshState() {
	// h.game.Mu.Lock()
	// defer h.game.Mu.Unlock()

	h.State.SmallBlindAmount = 0
	h.State.BigBlindAmount = 0
	h.State.CurrentRound = PreFlop
	h.State.Pot = 0
	h.State.CurrentBet = 0
	h.State.RoundComplete = false

	h.State.DealerSeat = nil
	h.State.CurrentSeat = nil
	h.State.SmallBlindSeat = nil
	h.State.BigBlindSeat = nil
	h.State.LastRaiserSeat = nil

	h.State.PlayerBets = make(map[string]int)
	h.State.Seats = make(map[int]*TableSeat)
	h.State.PlayerLastAction = make(map[string]HoldemActionType)
	h.State.CommunityCards = make([]models.Card, 0)
}

func (h *Holdem) ProcessAction(msg json.RawMessage) error {
	var action HoldemAction
	if err := json.Unmarshal(msg, &action); err != nil {
		return err
	}

	player := h.State.CurrentSeat.Player
	log.Printf("[INFO] Player %s processing action: %+v", player.Client.User.Player.ID, action)
	if action.PlayerID != player.Client.User.Player.ID {
		return errors.New("action received from wrong player")
	}

	if action.Action == HoldemActionRaise && h.State.CurrentBet == 0 {
		action.Action = HoldemActionBet
	} else if action.Action == HoldemActionBet && h.State.CurrentBet > 0 {
		action.Action = HoldemActionRaise
	}

	playerBet := h.State.PlayerBets[player.Client.User.Player.ID]
	toCall := min(h.State.CurrentBet-playerBet, player.Balance)

	switch action.Action {
	case HoldemActionFold:
		h.State.CurrentSeat.Hand = nil
		h.State.PlayerLastAction[player.Client.User.Player.ID] = HoldemActionFold
		log.Printf("[INFO] Player %s folds", player.Client.User.Player.ID)

	case HoldemActionCheck:
		if toCall > 0 {
			return errors.New("cannot check, must call or raise")
		}
		h.State.PlayerBets[player.Client.User.Player.ID] += 0
		h.State.PlayerLastAction[player.Client.User.Player.ID] = HoldemActionCheck
		log.Printf("[INFO] Player %s checks", player.Client.User.Player.ID)

	case HoldemActionCall:
		if toCall == 0 {
			return errors.New("nothing to call, must check")
		}
		player.Balance -= toCall
		h.State.Pot += toCall
		h.State.PlayerBets[player.Client.User.Player.ID] += toCall
		h.State.PlayerLastAction[player.Client.User.Player.ID] = HoldemActionCall
		log.Printf("[INFO] Player %s calls %d", player.Client.User.Player.ID, toCall)

	case HoldemActionBet:
		if h.State.CurrentBet > 0 {
			return errors.New("cannot bet, must raise")
		}
		if action.Amount < h.State.BigBlindAmount {
			return errors.New("bet must be at least the big blind")
		}
		if action.Amount > player.Balance {
			return errors.New("insufficient balance to bet")
		}

		player.Balance -= action.Amount
		h.State.Pot += action.Amount
		h.State.PlayerBets[player.Client.User.Player.ID] += action.Amount
		h.State.CurrentBet = action.Amount
		h.State.LastRaiserSeat = h.State.CurrentSeat
		h.State.PlayerLastAction[player.Client.User.Player.ID] = HoldemActionBet

		log.Printf("[INFO] Player %s bets %d", player.Client.User.Player.ID, action.Amount)

	case HoldemActionRaise:
		if h.State.CurrentBet == 0 {
			return errors.New("cannot raise, must bet")
		}
		minRaise := h.State.CurrentBet*2 - playerBet
		if action.Amount < minRaise {
			return errors.New("raise must be at least double the current bet")
		}
		if action.Amount > player.Balance {
			return errors.New("insufficient balance to raise")
		}

		toRaise := action.Amount - playerBet
		player.Balance -= toRaise
		h.State.Pot += toRaise
		h.State.PlayerBets[player.Client.User.Player.ID] += toRaise
		h.State.CurrentBet = action.Amount
		h.State.LastRaiserSeat = h.State.CurrentSeat
		h.State.PlayerLastAction[player.Client.User.Player.ID] = HoldemActionRaise
		log.Printf("[INFO] Player %s raises to %d", player.Client.User.Player.ID, action.Amount)

	case HoldemActionAllIn:
		if player.Balance == 0 {
			return errors.New("player already all-in")
		}

		allInAmount := player.Balance + playerBet
		h.State.Pot += player.Balance
		h.State.PlayerBets[player.Client.User.Player.ID] += player.Balance

		if allInAmount > h.State.CurrentBet {
			h.State.CurrentBet = allInAmount
			h.State.LastRaiserSeat = h.State.CurrentSeat
		}

		player.Balance = 0
		h.State.PlayerLastAction[player.Client.User.Player.ID] = HoldemActionAllIn
		log.Printf("[INFO] Player %s goes all-in with %d", player.Client.User.Player.ID, allInAmount)
	}

	h.SendMessage(HoldemMessagePlayerAction, HoldemPlayerActionMessage{
		PlayerID:  player.Client.User.Player.ID,
		Action:    action.Action,
		Amount:    action.Amount,
		GameState: h.GetGameState(),
	})

	h.LogGameState(fmt.Sprintf("AFTER %s ACTION BY %s", action.Action, player.Client.User.Player.ID))

	return nil
}

func (h *Holdem) DealCards() error {
	for _, seat := range h.State.Seats {
		seat.Hand = seat.Hand[:0]
	}

	currentSeat := h.State.DealerSeat.Next
	startPos := currentSeat.Position
	for {
		if currentSeat.Player.Status == GamePlayerStatusActive {
			card, err := h.deck.Draw()
			if err != nil {
				return err
			}
			currentSeat.Hand = append(currentSeat.Hand, card)
		}
		currentSeat = currentSeat.Next
		if currentSeat.Position == startPos {
			break
		}
	}

	currentSeat = h.State.DealerSeat.Next
	for {
		if currentSeat.Player.Status == GamePlayerStatusActive {
			card, err := h.deck.Draw()
			if err != nil {
				return err
			}
			currentSeat.Hand = append(currentSeat.Hand, card)
			log.Printf("[INFO] Dealt cards to player %s: %v",
				currentSeat.Player.Client.User.Player.ID, currentSeat.Hand)
		}
		currentSeat = currentSeat.Next
		if currentSeat.Position == startPos {
			break
		}
	}

	if err := h.DealCommunityCards(); err != nil {
		return err
	}

	return nil
}

func (h *Holdem) PlayRound() {
	log.Printf("[INFO] Starting new round")
	h.HandlePlayers()
	log.Printf("[INFO] Players after handling: %v", &h.game.Players)
	h.deck = models.NewDeck()

	h.StartPreFlopRound()
	if !h.CanGameContinue() {
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
	if !h.CanGameContinue() {
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
	if !h.CanGameContinue() {
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

	timer := time.NewTimer(1 * time.Second)
	<-timer.C

	if h.CanGameContinue() {
		h.PlayRound()
	} else {
		h.End()
	}
}

func (h *Holdem) StartPreFlopRound() {
	h.State.CurrentRound = PreFlop
	h.State.RoundComplete = false
	h.State.Pot = 0
	h.State.CurrentBet = 0
	h.State.LastRaiserSeat = nil
	h.State.PlayerBets = make(map[string]int)
	h.State.PlayerLastAction = make(map[string]HoldemActionType)

	h.RotateDealerButton()
	if !h.SetBlindPositions() {
		log.Printf("[ERROR] Not enough players to set blind positions")
		return
	}

	if err := h.DealCards(); err != nil {
		log.Printf("[ERROR] Failed to deal cards: %v", err)
		return
	}

	h.PostBlinds()

	h.State.CurrentSeat = h.State.BigBlindSeat.Next
	h.LogGameState(fmt.Sprintf("HAND STARTED - PRE-FLOP BETTING BEGINS Small Blind: %d, Big Blind: %d", h.State.SmallBlindSeat.Position, h.State.BigBlindSeat.Position))
	if err := h.BettingRound(); err != nil {
		log.Printf("[ERROR] Betting round error: %v", err)
		return
	}
	h.LogGameState("PRE-FLOP BETTING COMPLETE")
}

func (h *Holdem) StartFlopRound(communityCards []models.Card) {
	h.State.CurrentRound = Flop
	h.State.RoundComplete = false
	h.State.CurrentBet = 0
	h.State.LastRaiserSeat = nil
	h.State.PlayerBets = make(map[string]int)
	h.State.PlayerLastAction = make(map[string]HoldemActionType)

	h.State.CommunityCards = communityCards[:3]
	h.State.CurrentSeat = h.State.DealerSeat.Next
	h.LogGameState("FLOP BETTING BEGINS")
	if err := h.BettingRound(); err != nil {
		log.Printf("[ERROR] Betting round error: %v", err)
		return
	}
	h.LogGameState("FLOP BETTING COMPLETE")
}

func (h *Holdem) StartTurnRound(communityCards []models.Card) {
	h.State.CurrentRound = Turn
	h.State.RoundComplete = false
	h.State.CurrentBet = 0
	h.State.LastRaiserSeat = nil
	h.State.PlayerBets = make(map[string]int)
	h.State.PlayerLastAction = make(map[string]HoldemActionType)
	h.State.CommunityCards = communityCards[:4]
	h.State.CurrentSeat = h.State.DealerSeat.Next
	h.LogGameState("TURN DEALT - TURN BETTING BEGINS")
	if err := h.BettingRound(); err != nil {
		log.Printf("[ERROR] Betting round error: %v", err)
		return
	}
	h.LogGameState("TURN BETTING COMPLETE")
}

func (h *Holdem) StartRiverRound(communityCards []models.Card) {
	h.State.CurrentRound = River
	h.State.RoundComplete = false
	h.State.CurrentBet = 0
	h.State.LastRaiserSeat = nil
	h.State.PlayerBets = make(map[string]int)
	h.State.PlayerLastAction = make(map[string]HoldemActionType)

	h.State.CommunityCards = communityCards[:5]
	h.State.CurrentSeat = h.State.DealerSeat.Next
	h.LogGameState("RIVER DEALT - RIVER BETTING BEGINS")
	if err := h.BettingRound(); err != nil {
		log.Printf("[ERROR] Betting round error: %v", err)
		return
	}
	h.LogGameState("RIVER BETTING COMPLETE")
}

func (h *Holdem) PostBlinds() {
	smallBlindAmount := min(h.State.SmallBlindAmount, h.State.SmallBlindSeat.Player.Balance)
	h.State.SmallBlindSeat.Player.Balance -= smallBlindAmount
	h.State.Pot += smallBlindAmount
	h.State.PlayerBets[h.State.SmallBlindSeat.Player.Client.User.Player.ID] = smallBlindAmount

	bigBlindAmount := min(h.State.BigBlindAmount, h.State.BigBlindSeat.Player.Balance)
	h.State.BigBlindSeat.Player.Balance -= bigBlindAmount
	h.State.Pot += bigBlindAmount
	h.State.CurrentBet = bigBlindAmount
	h.State.PlayerBets[h.State.BigBlindSeat.Player.Client.User.Player.ID] = bigBlindAmount

	log.Printf("[BLINDS] Player %s posts small blind %d", h.State.SmallBlindSeat.Player.Client.User.Player.ID, smallBlindAmount)
	log.Printf("[BLINDS] Player %s posts big blind: %d", h.State.BigBlindSeat.Player.Client.User.Player.ID, bigBlindAmount)
}

func (h *Holdem) BettingRound() error {
	log.Printf("[INFO] Starting betting round for %v", h.State.CurrentRound)
	for !h.State.RoundComplete {
		if h.State.CurrentSeat == nil {
			return errors.New("no current seat")
		}

		player := h.State.CurrentSeat.Player
		if h.State.CurrentSeat.Hand == nil || player.Balance == 0 { // Skip folded or all-in players
			h.State.CurrentSeat = h.State.CurrentSeat.Next
			continue
		}

		h.SendMessageToPlayer(player.Client.User.Player.ID, HoldemMessagePlayerTurn, HoldemPlayerTurnMessage{
			GameState: h.GetGameState(),
			Timeout:   10, // 5 seconds timeout for action
		})

		playerBet := h.State.PlayerBets[player.Client.User.Player.ID]
		toCall := h.State.CurrentBet - playerBet
		log.Printf("[ACTION] Player %s to act | Current bet: $%d | Player bet: $%d | To call: $%d | Balance: $%d", player.Client.User.Player.ID, h.State.CurrentBet, playerBet, toCall, player.Balance)

		timer := time.NewTimer(5 * time.Second)
		actionReceived := make(chan bool, 1)
		go func() {
			select {
			case action := <-h.actionsChannel:
				if action.PlayerID == player.Client.User.Player.ID {
					if err := h.ProcessAction(action.Data); err != nil {
						log.Printf("[ERROR] Failed to process action: %v", err)
						fold := HoldemAction{
							PlayerID: player.Client.User.Player.ID,
							Action:   HoldemActionFold,
						}
						foldData, _ := json.Marshal(fold)
						h.ProcessAction(foldData)
					}
					actionReceived <- true
				}
			case <-timer.C:
				log.Printf("[INFO] Player %s timed out", player.Client.User.Player.ID)
				h.handleTimeoutAction(player, toCall)
				actionReceived <- true
			}
		}()

		<-actionReceived
		timer.Stop()

		if h.CheckRoundComplete() {
			h.State.RoundComplete = true
			log.Printf("[DEBUG] Round completed naturally")
			h.SendMessage(HoldemMessageRoundEnd, h.GetGameState())
			break
		}

		h.State.CurrentSeat = h.State.CurrentSeat.Next
	}

	log.Printf("[INFO] Betting round complete for %v", h.State.CurrentRound)
	return nil
}

func (h *Holdem) handleTimeoutAction(player *GamePlayer, toCall int) {
	var autoAction HoldemAction
	random := rand.Intn(100)

	if toCall > 0 && toCall <= player.Balance {
		autoAction = HoldemAction{
			PlayerID: player.Client.User.Player.ID,
			Action:   HoldemActionCall,
			Amount:   toCall,
		}
	} else if random >= 50 {
		autoAction = HoldemAction{
			PlayerID: player.Client.User.Player.ID,
			Action:   HoldemActionRaise,
			Amount:   player.Balance * 2 / 100,
		}
	} else if toCall == 0 && random < 50 {
		autoAction = HoldemAction{
			PlayerID: player.Client.User.Player.ID,
			Action:   HoldemActionCheck,
		}
	} else {
		autoAction = HoldemAction{
			PlayerID: player.Client.User.Player.ID,
			Action:   HoldemActionFold,
		}
	}

	autoData, _ := json.Marshal(autoAction)
	h.ProcessAction(autoData)
}

func (h *Holdem) DealPlayerCards() error {
	h.deck.Shuffle()
	for pos := range h.State.Seats {
		h.State.Seats[pos].Hand = h.State.Seats[pos].Hand[:0]
	}

	startPos := h.State.DealerSeat.Next.Position
	for i := 0; i < len(h.State.Seats); i++ {
		playerPos := (startPos + i) % len(h.State.Seats)
		seat := h.State.Seats[playerPos]

		if seat.Player.Status == GamePlayerStatusActive {
			card, err := h.deck.Draw()
			if err != nil {
				return err
			}
			seat.Hand = append(seat.Hand, card)
		}
	}

	for i := 0; i < len(h.State.Seats); i++ {
		playerPos := (startPos + i) % len(h.State.Seats)
		seat := h.State.Seats[playerPos]

		if seat.Player.Status == GamePlayerStatusActive {
			card, err := h.deck.Draw()
			if err != nil {
				return err
			}
			seat.Hand = append(seat.Hand, card)
		}
	}

	for _, seat := range h.State.Seats {
		if seat.Player.Status == GamePlayerStatusActive {
			log.Printf("[INFO] Player %s received cards: %v", seat.Player.Client.User.Player.ID, seat.Hand)
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

	h.game.Players = slices.DeleteFunc(h.game.Players, func(p *GamePlayer) bool {
		return p.Status == GamePlayerStatusInactive || p.Balance < h.game.MinBet
	})

	for _, player := range h.game.Players {
		if player.Status == GamePlayerStatusWaiting {
			player.Status = GamePlayerStatusActive
		}

		_, ok := h.State.Seats[player.Position]
		if !ok {
			h.State.Seats[player.Position] = &TableSeat{
				Position: player.Position,
				Player:   player,
			}
		}

		h.State.Seats[player.Position].Hand = make([]models.Card, 0, 2)

	}

	h.LinkSeats()
}

func (h *Holdem) LinkSeats() {
	if len(h.State.Seats) == 0 {
		return
	}

	positions := make([]int, 0, len(h.State.Seats))
	for pos := range h.State.Seats {
		positions = append(positions, pos)
	}

	sort.Ints(positions)

	for i := 0; i < len(positions); i++ {
		position := positions[i]
		seat := h.State.Seats[position]
		nextIdx := (i + 1) % len(positions)
		prevIdx := (i - 1 + len(positions)) % len(positions)

		seat.Next = h.State.Seats[positions[nextIdx]]
		seat.Prev = h.State.Seats[positions[prevIdx]]
	}
}

func (h *Holdem) CheckRoundComplete() bool {
	if len(h.State.PlayerBets) == 0 {
		return false
	}

	highestBet := 0
	for _, bet := range h.State.PlayerBets {
		if bet > highestBet {
			highestBet = bet
		}
	}

	var startSeat *TableSeat           // Determine where to start checking from
	if h.State.LastRaiserSeat != nil { // If someone raised, start from the seat after them
		startSeat = h.State.LastRaiserSeat.Next
	} else {
		// If no one raised (everyone checked/called),
		// start from the seat after the big blind in pre-flop
		// or from the seat after the dealer in post-flop rounds
		if h.State.CurrentRound == PreFlop {
			startSeat = h.State.BigBlindSeat.Next
		} else {
			startSeat = h.State.DealerSeat.Next
		}
	}
	currentSeat := startSeat

	// If we have gone all the way around back to the startSeat, we are done
	for {
		// Skip folded or all-in players
		if currentSeat.Player != nil && currentSeat.Hand != nil && len(currentSeat.Hand) > 0 {
			playerBet, exists := h.State.PlayerBets[currentSeat.Player.Client.User.Player.ID]
			_, lastActionExists := h.State.PlayerLastAction[currentSeat.Player.Client.User.Player.ID]
			if !exists || playerBet < highestBet || !lastActionExists { // If player hasn't bet or their bet is less than the highest bet, round is not complete
				return false
			}
		}

		currentSeat = currentSeat.Next
		if currentSeat == startSeat { // If we've gone full circle and returned to starting point, we're done checking
			break
		}
	}

	// If we've made it here, all active players have either:
	// 1. Folded
	// 2. Gone all-in
	// 3. Called the highest bet
	return true
}

func (h *Holdem) CanGameContinue() bool {
	// Check if there are at least 2 players with chips
	playersWithChips := 0
	for _, player := range h.game.Players {
		if player.Status == GamePlayerStatusActive {
			playersWithChips++
		}
	}

	if playersWithChips < 2 {
		return false
	}

	if h.PlayersNotFoldedCount() <= 1 {
		return false
	}

	return true
}

func (h *Holdem) Start() error {
	if !h.CanStart() {
		return errors.New("not enough players to start")
	}

	h.game.Status = GameStatusStarting
	h.State.DealerSeat = nil
	log.Printf("[INFO] Starting holdem game")

	for _, player := range h.game.Players { // Initialize all waiting players as active with starting chips
		if player.Status == GamePlayerStatusWaiting {
			player.Status = GamePlayerStatusActive
		}
	}

	h.RefreshState()
	go h.StartMessageChannel()

	h.LogGameState("GAME STARTING")
	h.SendMessage(HoldemMessageGameStart, h.GetGameState())

	h.game.Status = GameStatusStarted
	go h.PlayRound()

	return nil
}

func (h *Holdem) End() error {
	h.game.Mu.Lock()
	defer h.game.Mu.Unlock()

	h.game.Players = slices.DeleteFunc(h.game.Players, func(p *GamePlayer) bool {
		return p.Status == GamePlayerStatusInactive || p.Balance < h.game.MinBet
	})

	h.game.Status = GameStatusWaiting
	log.Printf("[INFO] Ending holdem game")

	h.LogGameState("GAME ENDED")
	h.SendMessage(HoldemMessageGameEnd, h.GetGameState())
	time.Sleep(100 * time.Millisecond)
	h.doneChannel <- true

	return nil
}

func (h *Holdem) CanStart() bool {
	activePlayers := 0
	for _, player := range h.game.Players {
		if player.Status == GamePlayerStatusActive || player.Status == GamePlayerStatusWaiting {
			activePlayers++
		}
	}

	return activePlayers >= 2
}

func (h *Holdem) PlayersNotFoldedCount() int {
	count := 0
	for _, seat := range h.State.Seats {
		if seat.Hand != nil && len(seat.Hand) > 0 {
			count++
		}
	}

	return count
}

func (h *Holdem) OnPlayerJoin(player *GamePlayer) error {
	player.Status = GamePlayerStatusWaiting
	log.Printf("[INFO] Player %s joined the game", player.Client.User.Player.ID)

	if h.game.Status == GameStatusWaiting && h.CanStart() {
		err := h.Start()
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *Holdem) OnPlayerLeave(player *GamePlayer) error {
	// player.Status = GamePlayerStatusInactive
	log.Printf("[INFO] Player %s left the game", player.Client.User.Player.ID)

	return nil
}

func (h *Holdem) RotateDealerButton() {
	if len(h.State.Seats) == 0 {
		return
	}

	if h.State.DealerSeat == nil || h.game.Status == GameStatusStarting {
		// Find the lowest position number for initial dealer
		minPos := math.MaxInt32
		for pos := range h.State.Seats {
			if pos < minPos {
				minPos = pos
				h.State.DealerSeat = h.State.Seats[pos]
			}
		}
	} else {
		h.State.DealerSeat = h.State.DealerSeat.Next
	}
}

func (h *Holdem) SetBlindPositions() bool {
	if len(h.State.Seats) < 2 {
		return false
	}

	if len(h.State.Seats) == 2 { // Heads-up
		h.State.SmallBlindSeat = h.State.DealerSeat // In heads-up, dealer is SB
		h.State.BigBlindSeat = h.State.DealerSeat.Next
	} else {
		h.State.SmallBlindSeat = h.State.DealerSeat.Next
		h.State.BigBlindSeat = h.State.SmallBlindSeat.Next
	}

	return true
}

func (h *Holdem) EvaluateHands() error {
	log.Println("[INFO] Evaluating hands")
	activePlayers := []*TableSeat{}
	currentSeat := h.State.DealerSeat
	startPos := currentSeat.Position

	for {
		if currentSeat.Hand != nil && len(currentSeat.Hand) > 0 && currentSeat.Player.Status != GamePlayerStatusInactive {
			activePlayers = append(activePlayers, currentSeat)
		}
		currentSeat = currentSeat.Next
		if currentSeat.Position == startPos {
			break
		}
	}

	if len(activePlayers) == 0 {
		log.Println("[ERROR] No active players to evaluate hands")
		return nil
	}

	if len(activePlayers) == 1 {
		winner := activePlayers[0]
		winner.Player.Balance += h.State.Pot
		log.Printf("[INFO] Player %s wins %d (uncontested)", winner.Player.Client.User.Player.ID, h.State.Pot)

		h.SendMessage(HoldemMessageWinner, HoldemWinnerMessage{
			WinnerID: winner.Player.Client.User.Player.ID,
			Amount:   h.State.Pot,
			Reason:   "uncontested",
		})

		h.State.Pot = 0
		return nil
	}

	results := []HandResult{}
	for _, seat := range activePlayers {
		allCards := append([]models.Card{}, seat.Hand...)
		allCards = append(allCards, h.State.CommunityCards...)

		handRank, highCards := evaluateBestHand(allCards)
		results = append(results, HandResult{
			Rank:      handRank,
			HighCards: highCards,
			PlayerID:  seat.Player.Client.User.Player.ID,
		})

		log.Printf("[INFO] Player %s has %v (rank: %d)", seat.Player.Client.User.Player.ID, seat.Hand, handRank)
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
		for _, seat := range h.State.Seats {
			if seat.Player.Client.User.Player.ID == winner.PlayerID {
				seat.Player.Balance += winAmount
				if remaining > 0 && seat.Player.Client.User.Player.ID == winners[0].PlayerID {
					seat.Player.Balance += remaining
					remaining = 0
				}
				break
			}
		}
	}

	if len(winners) == 1 {
		log.Printf("[INFO] Player %s wins %d with %d", winners[0].PlayerID, h.State.Pot, winners[0].Rank)
	} else {
		winnerIDs := make([]string, len(winners))
		for i, winner := range winners {
			winnerIDs[i] = winner.PlayerID
		}
		log.Printf("[INFO] Players %v split the pot (%d each) with %d", winnerIDs, winAmount, winners[0].Rank)
	}

	h.SendMessage(HoldemMessageShowdown, HoldemShowdownMessage{
		Winners:   winners,
		Pot:       h.State.Pot,
		GameState: h.GetGameState(),
	})

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

	unique := []int{}
	seen := make(map[int]bool)

	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			unique = append(unique, value)
		}
	}

	if len(unique) > count {
		return unique[:count]
	}
	return unique
}

func (h *Holdem) SendMessage(msgType HoldemMessageType, data interface{}) {
	msg := HoldemMessage{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	h.messageChannel <- GameMessage{
		PlayerID:    "",
		MessageType: GameMessageTypePlayerAction,
		Data:        msg,
	}
}

func (h *Holdem) SendMessageToPlayer(playerID string, msgType HoldemMessageType, data interface{}) {
	msg := HoldemMessage{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	h.messageChannel <- GameMessage{
		PlayerID:    playerID,
		MessageType: GameMessageTypePlayerAction,
		Data:        msg,
	}
}

func (h *Holdem) StartMessageChannel() {
	for {
		select {
		case msg := <-h.messageChannel:
			if msg.ToGame {
				h.game.Room.SendMessageToRoom(msg)
			} else if msg.PlayerID == "" {
				h.game.Room.SendMessageToPlayer(msg.PlayerID, msg)
			} else {
				h.game.Room.SendMessageToRoom(msg)
			}
		case <-h.doneChannel:
			return
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (h *Holdem) GetPlayersInRound() []*GamePlayer {
	activePlayers := []*GamePlayer{}

	for _, player := range h.game.Players {
		if player.Status == GamePlayerStatusActive || player.Status == GamePlayerStatusInactive {
			activePlayers = append(activePlayers, player)
		}
	}

	return activePlayers
}

func (h *Holdem) GetActivePlayers() []*GamePlayer {
	activePlayers := []*GamePlayer{}

	for _, player := range h.game.Players {
		if player.Status == GamePlayerStatusActive {
			activePlayers = append(activePlayers, player)
		}
	}

	return activePlayers
}

type PlayerView struct {
	ID            string           `json:"id"`
	Status        GamePlayerStatus `json:"status"`
	Position      int              `json:"position"`
	Name          string           `json:"name"`
	Balance       int              `json:"balance"`
	Hand          []models.Card    `json:"hand"`
	IsFolded      bool             `json:"is_folded"`
	IsAllIn       bool             `json:"is_all_in"`
	IsDealer      bool             `json:"is_dealer"`
	IsSmallBlind  bool             `json:"is_small_blind"`
	IsBigBlind    bool             `json:"is_big_blind"`
	IsCurrentTurn bool             `json:"is_current_turn"`
}

type GameView struct {
	Players        []PlayerView
	CommunityCards []models.Card
	Pot            int
	CurrentBet     int
	CurrentRound   HoldemRound
}

func (h *Holdem) GetGameState() interface{} {
	visibleCards := []models.Card{}
	switch h.State.CurrentRound {
	case Flop:
		visibleCards = h.State.CommunityCards[:3]
	case Turn:
		visibleCards = h.State.CommunityCards[:4]
	case River, Showdown:
		visibleCards = h.State.CommunityCards[:5]
	}

	playerViews := []PlayerView{}
	for i, player := range h.game.Players {
		playerView := PlayerView{
			ID:       player.Client.User.Player.ID,
			Status:   player.Status,
			Position: player.Position,
			Name:     player.Client.User.Player.Username,
			Balance:  player.Balance,
			Hand:     []models.Card{},
			IsFolded: false,
			IsAllIn:  false,
		}

		seat, ok := h.State.Seats[player.Position]
		if !ok || h.game.Status != GameStatusStarted || h.State.DealerSeat == nil {
			playerViews = append(playerViews, playerView)
			continue
		}

		playerView.Hand = seat.Hand
		playerView.IsFolded = player.Status != GamePlayerStatusWaiting && seat.Hand == nil
		playerView.IsAllIn = player.Status != GamePlayerStatusWaiting && player.Balance == 0
		playerView.IsDealer = i == h.State.DealerSeat.Position
		playerView.IsSmallBlind = i == h.State.SmallBlindSeat.Position
		playerView.IsBigBlind = i == h.State.BigBlindSeat.Position
		playerView.IsCurrentTurn = i == h.State.CurrentSeat.Position
		playerViews = append(playerViews, playerView)
	}

	return GameView{
		Players:        playerViews,
		CommunityCards: visibleCards,
		Pot:            h.State.Pot,
		CurrentBet:     h.State.CurrentBet,
		CurrentRound:   h.State.CurrentRound,
	}
}

func (h *Holdem) GetPlayerState(playerID string) GameView {
	baseState := h.GetGameState()

	// Find the player's hand
	// var playerHand []models.Card
	for _, player := range h.game.Players {
		if player.Client.User.Player.ID == playerID {
			// playerHand = h.State.Seats[player.Position].Hand
			break
		}
	}

	// Add player's hand to the state
	// baseState.Players[playerID].Hand = playerHand

	return baseState.(GameView)
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
		seat, ok := h.State.Seats[player.Position]
		if ok {
			status := "Active"
			if seat.Hand == nil {
				status = "Folded"
			} else if player.Balance == 0 {
				status = "All-In"
			}
			if h.game.Status != GameStatusStarted || h.State.DealerSeat == nil {
				continue
			}

			position := ""
			if i == h.State.DealerSeat.Position {
				position += "Dealer "
			}
			if i == h.State.SmallBlindSeat.Position {
				position += "SB "
			}
			if i == h.State.BigBlindSeat.Position {
				position += "BB "
			}
			if i == h.State.CurrentSeat.Position {
				position += "Acting "
			}

			playerBet := h.State.PlayerBets[player.Client.User.Player.ID]

			log.Printf("  - Player %s (%s):", player.Client.User.Player.ID, position)
			log.Printf("	- Balance: $%d | Bet: $%d | Status: %s ", player.Balance, playerBet, status)
			log.Printf("	- Cards: %v", func() string {
				if seat.Hand == nil {
					return "Folded"
				}
				return fmt.Sprintf("%v", seat.Hand)
			}())
		} else {
			log.Printf("  - Player %s: (not in game)", player.Client.User.Player.ID)
		}
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
