package flashcard

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/avelino/slugify"
	"github.com/pelletier/go-toml"
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

func OpenDeck(path string, clock Clock) (Deck, error) {
	tree, err := toml.LoadFile(path)
	if err != nil {
		return Deck{}, fmt.Errorf("open deck file '%s' : %w", path, err)
	}

	var file deckFile
	if err := tree.Unmarshal(&file); err != nil {
		return Deck{}, fmt.Errorf("unmarshall deck '%s' : %w", path, err)
	}

	for id, card := range file.Cards {
		card.id = id
		file.Cards[id] = card
	}

	return Deck{Name: file.Name, cards: file.Cards, id: path, clock: clock}, nil
}

// Deck represents a named collection of cards.
type Deck struct {
	Name string

	cards map[string]Card
	id    string
	clock Clock
}

// List returns a collection of cards order by the time of the last review and question.
func (d Deck) List() []Card {
	cards := make([]Card, 0, len(d.cards))
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
func (d Deck) DueCards() []Card {
	cards := make([]Card, 0, len(d.cards))

	if d.clock == nil {
		return cards
	}

	date := d.clock.Now()
	for _, card := range d.cards {
		if card.Due(date) {
			cards = append(cards, card)
		}
	}

	return cards
}

// HasDueCards says if the deck has due cards.
func (d Deck) HasDueCards() bool {
	return len(d.DueCards()) > 0
}

// Id returns the deck identifier.
func (d Deck) Id() string {
	return d.id
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

// NewDeckRepository create a new deck repository by reading all decks
// from a given folder.
func NewDeckRepository(directory string, clock Clock) (*DeckRepository, error) {
	if err := assureDirExist(directory); err != nil {
		return nil, err
	}

	decks, err := loadDecks(directory, clock)
	if err != nil {
		return nil, err
	}
	return &DeckRepository{decks: decks, clock: clock, directory: directory}, nil
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

func loadDecks(path string, clock Clock) (map[string]Deck, error) {
	files, err := filepath.Glob(path + "/*.toml")
	if err != nil {
		return nil, fmt.Errorf("find decks: %w", err)
	}

	decks := make(map[string]Deck, len(files))
	for _, file := range files {
		deck, err := OpenDeck(file, clock)
		if err != nil {
			return nil, err
		}
		decks[deck.id] = deck
	}
	return decks, nil
}

// DeckRepository defines storage interface that manages a set of decks.
type DeckRepository struct {
	directory string
	decks     map[string]Deck
	clock     Clock
}

// List returns the available deck names.
func (r *DeckRepository) List() []Deck {
	decks := make([]Deck, 0, len(r.decks))
	for _, deck := range r.decks {
		decks = append(decks, deck)
	}

	sort.Slice(decks, func(i, j int) bool {
		return decks[i].Name < decks[j].Name
	})

	return decks
}

// Total returns the number of decks.
func (r *DeckRepository) Total() int {
	return len(r.decks)
}

// Create creates a new deck from a given name.
func (r *DeckRepository) Create(name string) (Deck, error) {
	deck := Deck{
		Name:  name,
		cards: make(map[string]Card),
		clock: r.clock,
		id:    r.deckPath(name),
	}

	r.decks[deck.id] = deck
	if err := r.Save(deck); err != nil {
		delete(r.decks, deck.id)
		return Deck{}, err
	}

	return deck, nil
}

func (r *DeckRepository) deckPath(name string) string {
	return filepath.Join(r.directory, slugify.Slugify(name)+".toml")
}

// Open returns a deck given a name.
func (r DeckRepository) Open(name string) (Deck, error) {
	for _, deck := range r.decks {
		if deck.Name == name {
			return deck, nil
		}
	}
	return Deck{}, ErrDeckNotExist
}

// Save writes changes to disk.
func (r *DeckRepository) Save(deck Deck) error {
	if _, ok := r.decks[deck.id]; !ok {
		return ErrDeckNotExist
	}

	data, err := toml.Marshal(deckFile{deck.Name, deck.cards})
	if err != nil {
		return fmt.Errorf("marshall deck: %w", err)
	}

	if err := os.WriteFile(deck.id, data, 0644); err != nil {
		return fmt.Errorf("write deck: %w", err)
	}

	return nil
}

// Remove removes the deck from the repository.
func (r *DeckRepository) Remove(deck Deck) error {
	if _, ok := r.decks[deck.id]; !ok {
		return ErrDeckNotExist
	}

	if err := os.Remove(deck.id); err != nil {
		return fmt.Errorf("remove deck '%s': %w", deck.Name, err)
	}

	delete(r.decks, deck.id)

	return nil
}

// SaveStats writes stats to disk.
func (r *DeckRepository) SaveStats(deck Deck, stats *Stats) error {
	data, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("marshal stats: %w", err)
	}

	f, err := os.OpenFile(r.statsPath(deck), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open stats file: %w", err)
	}

	_, err = f.Write(append(data, '\n'))
	if err0 := f.Close(); err0 != nil && err == nil {
		err = fmt.Errorf("%w ", err)
	}

	if err != nil {
		return fmt.Errorf("write stats: %w", err)
	}

	return nil
}

func (r *DeckRepository) statsPath(d Deck) string {
	return filepath.Join(r.directory, slugify.Slugify(d.Name)+"-stats.jsonl")
}
