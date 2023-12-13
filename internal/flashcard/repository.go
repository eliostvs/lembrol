package flashcard

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/avelino/slugify"
	"github.com/go-playground/validator/v10"

	"github.com/eliostvs/lembrol/internal/clock"
)

var ErrDeckNotFound = errors.New("deck not found")

// NewRepository create a new deck repository by reading all decks
// from a given folder.
func NewRepository(path string, clock clock.Clock) (*Repository, error) {
	if err := assureDirectoryExist(path); err != nil {
		return nil, err
	}

	decks, err := loadDecks(path, clock)
	if err != nil {
		return nil, err
	}
	return &Repository{decks: decks, clock: clock, path: path, validator: validator.New()}, nil
}

func assureDirectoryExist(path string) error {
	_, err := os.Stat(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err != nil {
		return os.MkdirAll(path, 0o777)
	}
	return nil
}

func loadDecks(path string, clock clock.Clock) (map[string]Deck, error) {
	files, err := filepath.Glob(path + "/*.json")
	if err != nil {
		return nil, fmt.Errorf("reading deck '%s': %w", path, err)
	}

	decks := make(map[string]Deck, len(files))
	for _, file := range files {
		deck, err := openDeck(file, clock)
		if err != nil {
			return nil, err
		}
		decks[deck.ID] = deck
	}
	return decks, nil
}

func openDeck(filename string, clock clock.Clock) (Deck, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return Deck{}, fmt.Errorf("read deck file '%s' : %w", filename, err)
	}

	var deck Deck
	if err := json.Unmarshal(data, &deck); err != nil {
		return Deck{}, fmt.Errorf("unmarshall deck '%s' : %w", filename, err)
	}

	deck.ID = filepathBaseWithoutExt(filename)
	deck.clock = clock

	return deck, nil
}

func filepathBaseWithoutExt(filename string) string {
	return strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
}

// Repository defines storage interface that manages a set of decks.
type Repository struct {
	path      string
	decks     map[string]Deck
	clock     clock.Clock
	validator *validator.Validate
}

// List returns the available deck names.
func (r *Repository) List() []Deck {
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
func (r *Repository) Total() int {
	return len(r.decks)
}

// Create creates a new deck from a given name.
func (r *Repository) Create(name string, cards []Card) (Deck, error) {
	deck, err := NewDeck(name, r.clock, cards)
	if err != nil {
		return deck, err
	}

	if _, ok := r.decks[deck.ID]; ok {
		return Deck{}, fmt.Errorf("deck '%s' already exists", deck.Name)
	}

	if err := r.Save(deck); err != nil {
		return Deck{}, err
	}
	r.decks[deck.ID] = deck

	return deck, nil
}

// Save writes changes to disk.
func (r *Repository) Save(deck Deck) error {
	if err := r.validator.Struct(deck); err != nil {
		return fmt.Errorf("failed to validate: %w", err)
	}

	data, err := json.Marshal(&deck)
	if err != nil {
		return fmt.Errorf("failed to marshal deck: %w", err)
	}

	if err := os.WriteFile(deckFilepath(r.path, deck.ID), data, 0o644); err != nil {
		return fmt.Errorf("write deck: %w", err)
	}

	delete(r.decks, deck.ID)
	r.decks[deck.ID] = deck

	return nil
}

// Delete removes the deck from the repository.
func (r *Repository) Delete(deck Deck) error {
	if _, ok := r.decks[deck.ID]; !ok {
		return ErrDeckNotFound
	}

	if err := os.Remove(deckFilepath(r.path, deck.ID)); err != nil {
		return fmt.Errorf("delete deck '%s': %w", deck.ID, err)
	}

	delete(r.decks, deck.ID)

	return nil
}

// Find searches deck by name.
func (r *Repository) Find(name string) (Deck, error) {
	for _, deck := range r.decks {
		if strings.EqualFold(deck.Name, name) {
			return deck, nil
		}
	}

	return Deck{}, ErrDeckNotFound
}

func deckFilepath(dirname, filename string) string {
	return filepath.Join(dirname, slugify.Slugify(filename)+".json")
}
