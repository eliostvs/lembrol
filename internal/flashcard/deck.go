package flashcard

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/avelino/slugify"
	"github.com/pelletier/go-toml"

	"github.com/eliostvs/lembrol/internal/clock"
)

var (
	// ErrDeckNotExist indicates deck does not exist.
	ErrDeckNotExist = errors.New("deck not exist")
	// ErrCardNotExist indicates deck does not exist.
	ErrCardNotExist = errors.New("card not exist")
)

type deckFile struct {
	Name  string
	Cards map[string]Card
}

func ReadDeck(location string, clock clock.Clock) (Deck, error) {
	tree, err := toml.LoadFile(location)
	if err != nil {
		return Deck{}, fmt.Errorf("open deck file '%s' : %w", location, err)
	}

	var d deckFile
	if err := tree.Unmarshal(&d); err != nil {
		return Deck{}, fmt.Errorf("unmarshall deck '%s' : %w", location, err)
	}

	for id, card := range d.Cards {
		card.id = id
		d.Cards[id] = card
	}

	return newDeck(d.Name, location, clock, d.Cards), nil
}

func newDeck(name, location string, clock clock.Clock, cards map[string]Card) Deck {
	if cards == nil {
		cards = make(map[string]Card)
	}

	return Deck{
		Name:     name,
		location: location,
		cards:    cards,
		clock:    clock,
	}
}

// Deck represents a named collection of cards.
type Deck struct {
	Name string

	location string
	cards    map[string]Card
	clock    clock.Clock
}

// List returns a collection of cards order by the time of the last review and question.
func (d Deck) List() []Card {
	cards := make([]Card, 0, len(d.cards))
	for _, card := range d.cards {
		cards = append(cards, card)
	}

	sort.Slice(
		cards, func(i, j int) bool {
			if cards[i].ReviewedAt.Equal(cards[j].ReviewedAt) {
				return cards[i].Question < cards[j].Question
			}
			return cards[i].ReviewedAt.After(cards[j].ReviewedAt)
		},
	)

	return cards
}

// DueCards returns a collection of cards that needs review.
func (d Deck) DueCards() []Card {
	cards := make([]Card, 0, len(d.cards))

	if d.clock == nil {
		return cards
	}

	date := d.clock.Now()
	for _, card := range d.cards {
		if card.IsDue(date) {
			cards = append(cards, card)
		}
	}

	return cards
}

// HasDueCards says if the deck has due cards.
func (d Deck) HasDueCards() bool {
	return len(d.DueCards()) > 0
}

// Total returns number of cards in the deck.
func (d Deck) Total() int {
	return len(d.cards)
}

// Add adds a new card to the deck.
func (d Deck) Add(question, answer string) (Deck, Card) {
	card := NewCard(question, answer, d.clock.Now())
	d.cards[card.id] = card
	return d, card
}

// Change updates a card in the deck.
func (d Deck) Change(card Card) Deck {
	d.cards[card.id] = card
	return d
}

// Remove excludes card from the deck.
func (d Deck) Remove(card Card) (Deck, error) {
	if _, ok := d.cards[card.id]; !ok {
		return Deck{}, ErrCardNotExist
	}

	delete(d.cards, card.id)
	return d, nil
}

func (d Deck) Validate() error {
	if d.Name == "" {
		return errors.New("invalid empty file name")
	}
	return nil
}

// NewDeckRepository create a new deck repository by reading all decks
// from a given folder.
func NewDeckRepository(location string, clock clock.Clock) (*DeckRepository, error) {
	if err := assureDirExist(location); err != nil {
		return nil, err
	}

	decks, err := loadDecks(location, clock)
	if err != nil {
		return nil, err
	}
	return &DeckRepository{decks: decks, clock: clock, location: location}, nil
}

func assureDirExist(path string) error {
	_, err := os.Stat(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err != nil {
		return os.MkdirAll(path, 0777)
	}
	return nil
}

func loadDecks(path string, clock clock.Clock) (map[string]Deck, error) {
	files, err := filepath.Glob(path + "/*.toml")
	if err != nil {
		return nil, fmt.Errorf("find decks: %w", err)
	}

	decks := make(map[string]Deck, len(files))
	for _, file := range files {
		deck, err := ReadDeck(file, clock)
		if err != nil {
			return nil, err
		}
		decks[deck.location] = deck
	}
	return decks, nil
}

// DeckRepository defines storage interface that manages a set of decks.
type DeckRepository struct {
	location string
	decks    map[string]Deck
	clock    clock.Clock
}

// List returns the available deck names.
func (r *DeckRepository) List() []Deck {
	decks := make([]Deck, 0, len(r.decks))
	for _, deck := range r.decks {
		decks = append(decks, deck)
	}

	sort.Slice(
		decks, func(i, j int) bool {
			return decks[i].Name < decks[j].Name
		},
	)

	return decks
}

// Total returns the number of decks.
func (r *DeckRepository) Total() int {
	return len(r.decks)
}

// Create creates a new deck from a given name.
func (r *DeckRepository) Create(name string) (Deck, error) {
	deck := newDeck(name, r.path(name), r.clock, nil)

	if err := r.Save(deck); err != nil {
		return Deck{}, err
	}
	r.decks[deck.location] = deck

	return deck, nil
}

func (r *DeckRepository) path(name string) string {
	return filepath.Join(r.location, slugify.Slugify(name)+".toml")
}

// Open returns a deck given a name.
func (r *DeckRepository) Open(name string) (Deck, error) {
	for _, deck := range r.decks {
		if deck.Name == name {
			return deck, nil
		}
	}
	return Deck{}, ErrDeckNotExist
}

// Save writes changes to disk.
func (r *DeckRepository) Save(deck Deck) error {
	if err := deck.Validate(); err != nil {
		return err
	}

	data, err := toml.Marshal(deckFile{deck.Name, deck.cards})
	if err != nil {
		return fmt.Errorf("marshall deck: %w", err)
	}

	if err := os.WriteFile(deck.location, data, 0644); err != nil {
		return fmt.Errorf("write deck: %w", err)
	}

	return nil
}

// Remove removes the deck from the repository.
func (r *DeckRepository) Remove(deck Deck) error {
	if _, ok := r.decks[deck.location]; !ok {
		return ErrDeckNotExist
	}

	if err := os.Remove(deck.location); err != nil {
		return fmt.Errorf("remove deck '%s': %w", deck.Name, err)
	}

	delete(r.decks, deck.location)

	return nil
}
