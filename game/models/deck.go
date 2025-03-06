package models

import (
	"errors"
	"math/rand"
	"time"
)

var (
	ErrEmptyDeck = errors.New("deck is empty")
)

// Deck represents a standard deck of 52 playing cards
type Deck struct {
	Cards []Card
}

type Card struct {
	Suit   string `json:"suit"`
	Value  int    `json:"value"`
	Hidden bool   `json:"hidden"`
}

// Suit represents a card suit
type Suit string

const (
	Hearts   Suit = "hearts"
	Diamonds Suit = "diamonds"
	Clubs    Suit = "clubs"
	Spades   Suit = "spades"
)

// Value represents a card value
type Value int

const (
	Two   Value = 2
	Three Value = 3
	Four  Value = 4
	Five  Value = 5
	Six   Value = 6
	Seven Value = 7
	Eight Value = 8
	Nine  Value = 9
	Ten   Value = 10
	Jack  Value = 11
	Queen Value = 12
	King  Value = 13
	Ace   Value = 14
)

// NewDeck creates and returns a new deck of cards
func NewDeck() *Deck {
	deck := &Deck{Cards: make([]Card, 0, 52)}
	suits := []Suit{Hearts, Diamonds, Clubs, Spades}
	values := []Value{Two, Three, Four, Five, Six, Seven, Eight, Nine, Ten, Jack, Queen, King, Ace}

	for _, suit := range suits {
		for _, value := range values {
			deck.Cards = append(deck.Cards, Card{
				Suit:   string(suit),
				Value:  int(value),
				Hidden: true,
			})
		}
	}

	deck.Shuffle()
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
