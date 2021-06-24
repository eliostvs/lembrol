package flashcard

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/avelino/slugify"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/pelletier/go-toml"
)

func OpenDeck(filename string, clock Clock) (*Deck, error) {
	tree, err := toml.LoadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("open deck file '%s' : %w", filename, err)
	}

	var d file
	if err := tree.Unmarshal(&d); err != nil {
		return nil, fmt.Errorf("unmarshall deck '%s' : %w", filename, err)
	}

	cards := make(map[string]*Card, len(d.Cards))
	for _, card := range d.Cards {
		card.id = gonanoid.Must()
		cards[card.id] = card
	}

	return &Deck{filename: filename, clock: clock, id: gonanoid.Must(), cards: cards, Name: d.Name}, nil
}

func newDeck(name, dir string, clock Clock) *Deck {
	return &Deck{
		id:       gonanoid.Must(),
		Name:     name,
		filename: filepath.Join(dir, slugify.Slugify(name)+".toml"),
		clock:    clock,
		cards:    make(map[string]*Card),
	}
}

// Deck represents a named collection of cards.
type Deck struct {
	Name string

	cards    map[string]*Card
	id       string
	filename string
	clock    Clock
}

// List returns a collection of cards order by the time of the last review and question.
func (d *Deck) List() []*Card {
	cards := make([]*Card, 0, len(d.cards))
	for _, card := range d.cards {
		cards = append(cards, card)
	}

	sort.Slice(cards, func(i, j int) bool {
		if cards[i].ReviewedAt.Equal(cards[j].ReviewedAt) {
			return cards[i].Question < cards[j].Question
		}
		return cards[i].ReviewedAt.After(cards[j].ReviewedAt)
	})

	return cards
}

// DueCards returns a collection of cards that needs review.
func (d *Deck) DueCards() []*Card {
	date := d.clock.Now()

	dueCards := make([]*Card, 0, len(d.cards))
	for _, card := range d.cards {
		if card.Due(date) {
			dueCards = append(dueCards, card)
		}
	}

	return dueCards
}

// Remove excludes card from the deck.
func (d *Deck) Remove(card *Card) error {
	if card == nil {
		return ErrCardNotExist
	}

	if _, ok := d.cards[card.id]; !ok {
		return ErrCardNotExist
	}

	delete(d.cards, card.id)
	return nil
}

// HasDueCards says if the deck has due cards.
func (d *Deck) HasDueCards() bool {
	return len(d.DueCards()) > 0
}

// Total returns number of cards in the deck.
func (d *Deck) Total() int {
	return len(d.cards)
}

// Add adds a new card to the deck.
func (d *Deck) Add(question, answer string) *Card {
	card := NewCard(gonanoid.Must(), question, answer, d.clock.Now())
	d.cards[card.id] = card
	return card
}

// Id returns the deck identifier.
func (d *Deck) Id() string {
	return d.id
}
