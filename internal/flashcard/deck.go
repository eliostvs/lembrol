package flashcard

import (
	"errors"
	"sort"

	"github.com/avelino/slugify"

	"github.com/eliostvs/lembrol/internal/clock"
)

// NewDeck creates a new deck.
func NewDeck(name string, clock clock.Clock, cards []Card) (deck Deck, err error) {
	if clock == nil {
		return deck, errors.New("missing clock")
	}

	if name == "" {
		return deck, errors.New("missing name")
	}

	deck = Deck{
		ID:    slugify.Slugify(name),
		Name:  name,
		Cards: cards,
		clock: clock,
	}

	return deck, nil
}

// Deck represents a named collection of cards.
type Deck struct {
	Name  string `json:"name" validate:"required"`
	Cards []Card `json:"cards"`

	ID    string
	clock clock.Clock
}

// List returns a collection of cards order by the time of the last review and question.
func (d Deck) List() []Card {
	cards := make([]Card, 0, len(d.Cards))
	cards = append(cards, d.Cards...)
	d.sort(cards)
	return cards
}

func (d Deck) sort(cards []Card) {
	sort.Slice(
		cards, func(i, j int) bool {
			if cards[i].LastReview.Equal(cards[j].LastReview) {
				return cards[i].Question < cards[j].Question
			}
			return cards[i].LastReview.After(cards[j].LastReview)
		},
	)
}

// DueCards returns a collection of cards that needs review.
func (d Deck) DueCards() []Card {
	cards := make([]Card, 0, len(d.Cards))

	// avoid panic when deck was not initialized correctly
	if d.clock == nil {
		return cards
	}

	date := d.clock.Now()
	for _, card := range d.Cards {
		if card.IsDue(date) {
			cards = append(cards, card)
		}
	}

	d.sort(cards)

	return cards
}

// HasDueCards says if the deck has due cards.
func (d Deck) HasDueCards() bool {
	return len(d.DueCards()) > 0
}

// Total returns number of cards in the deck.
func (d Deck) Total() int {
	return len(d.Cards)
}

// Add adds a new card to the deck.
func (d Deck) Add(question, answer string) (Deck, Card) {
	card := NewCard(question, answer, d.clock.Now())
	d.Cards = append(d.Cards, card)
	return d, card
}

// Change updates a card in the deck.
func (d Deck) Change(updated Card) Deck {
	cards := make([]Card, 0, len(d.Cards))
	for _, card := range d.Cards {
		if card.ID == updated.ID {
			cards = append(cards, updated)
		} else {
			cards = append(cards, card)
		}
	}
	d.Cards = cards
	return d
}

// Remove excludes card from the deck.
func (d Deck) Remove(card Card) Deck {
	cards := make([]Card, 0, len(d.Cards))

	for _, c := range d.Cards {
		if c.ID != card.ID {
			cards = append(cards, c)
		}
	}

	d.Cards = cards
	return d
}
