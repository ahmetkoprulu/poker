package models

import (
	"errors"
	"math/rand"
	"time"
)

var (
	ErrEmptyDeck = errors.New("deck is empty")
)

// Suit represents a card suit
type Suit string

const (
	Hearts   Suit = "hearts"
	Diamonds Suit = "diamonds"
	Clubs    Suit = "clubs"
	Spades   Suit = "spades"
)

// Value represents a card value
type Value string

const (
	Two   Value = "2"
	Three Value = "3"
	Four  Value = "4"
	Five  Value = "5"
	Six   Value = "6"
	Seven Value = "7"
	Eight Value = "8"
	Nine  Value = "9"
	Ten   Value = "10"
	Jack  Value = "J"
	Queen Value = "Q"
	King  Value = "K"
	Ace   Value = "A"
)

// Round represents a betting round in poker
type Round string

const (
	PreFlop Round = "preflop"
	Flop    Round = "flop"
	Turn    Round = "turn"
	River   Round = "river"
)

// HandRank represents the rank of a poker hand
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

// Deck represents a standard deck of 52 playing cards
type Deck struct {
	Cards []Card
}

// NewDeck creates and returns a new deck of cards
func NewDeck() *Deck {
	deck := &Deck{Cards: make([]Card, 0, 52)}
	suits := []Suit{Hearts, Diamonds, Clubs, Spades}
	values := []Value{Two, Three, Four, Five, Six, Seven, Eight, Nine, Ten, Jack, Queen, King, Ace}

	for _, suit := range suits {
		for _, value := range values {
			deck.Cards = append(deck.Cards, Card{
				Suit:   string(suit),
				Value:  string(value),
				Hidden: true,
			})
		}
	}

	return deck
}

// Shuffle randomizes the order of cards in the deck
func (d *Deck) Shuffle() {
	rand.Seed(time.Now().UnixNano())
	for i := len(d.Cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	}
}

// Draw removes and returns the top card from the deck
func (d *Deck) Draw() (Card, error) {
	if len(d.Cards) == 0 {
		return Card{}, ErrEmptyDeck
	}

	card := d.Cards[len(d.Cards)-1]
	d.Cards = d.Cards[:len(d.Cards)-1]
	return card, nil
}

// PokerHand represents a player's poker hand with its rank
type PokerHand struct {
	Cards []Card
	Rank  HandRank
	Value int // Used for comparing hands of the same rank
}

// EvaluateHand determines the best poker hand from the given cards
func EvaluateHand(playerCards []Card, communityCards []Card) PokerHand {
	// Combine player's cards with community cards
	allCards := append([]Card{}, playerCards...)
	allCards = append(allCards, communityCards...)

	// TODO: Implement hand evaluation logic
	// This will be a complex function that needs to:
	// 1. Check for all possible hand combinations
	// 2. Return the best possible hand
	// 3. Assign proper rank and value for comparison

	return PokerHand{
		Cards: allCards,
		Rank:  HighCard,
		Value: 0,
	}
}

// BettingRound manages a single round of betting
type BettingRound struct {
	Round      Round
	CurrentBet int
	Pot        int
	Actions    []GameAction
}

// NewBettingRound creates a new betting round
func NewBettingRound(round Round, currentBet int, pot int) *BettingRound {
	return &BettingRound{
		Round:      round,
		CurrentBet: currentBet,
		Pot:        pot,
		Actions:    make([]GameAction, 0),
	}
}
