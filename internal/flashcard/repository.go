package flashcard

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/pelletier/go-toml"
)

var (
	// ErrDeckNotExist indicates deck does not exist.
	ErrDeckNotExist = errors.New("deck not exist")
	// ErrCardNotExist indicates deck does not exist.
	ErrCardNotExist = errors.New("card not exist")
)

type file struct {
	Name  string
	Cards []Card
}

// NewRepository create a new Repository by reading all decks
// from a given folder.
func NewRepository(path string, clock Clock) (*Repository, error) {
	if err := assureDirExist(path); err != nil {
		return nil, err
	}

	decks, err := loadDecks(path, clock)
	if err != nil {
		return nil, err
	}
	return &Repository{decks: decks, clock: clock, path: path}, nil
}

// Repository defines storage interface that manages a set of decks.
type Repository struct {
	path  string
	decks map[string]*Deck
	clock Clock
}

// Open returns a deck given a name.
func (r Repository) Open(name string) (*Deck, error) {
	for _, deck := range r.decks {
		if deck.Name == name {
			return deck, nil
		}
	}
	return nil, ErrDeckNotExist
}

// List returns the available deck names.
func (r *Repository) List() []*Deck {
	decks := make([]*Deck, 0, len(r.decks))
	for _, deck := range r.decks {
		decks = append(decks, deck)
	}

	sort.Slice(decks, func(i, j int) bool {
		return decks[i].Name < decks[j].Name
	})

	return decks
}

// Total returns the number of decks.
func (r *Repository) Total() int {
	return len(r.decks)
}

// Create creates a new deck from a given name.
func (r *Repository) Create(name string) (*Deck, error) {
	deck := newDeck(name, r.path, r.clock)

	r.decks[deck.id] = deck
	if err := r.Save(deck); err != nil {
		delete(r.decks, deck.id)
		return nil, err
	}

	return deck, nil
}

// Save persists the changes in a deck.
func (r *Repository) Save(deck *Deck) error {
	if deck == nil {
		return ErrDeckNotExist
	}

	if _, ok := r.decks[deck.id]; !ok {
		return ErrDeckNotExist
	}

	data, err := toml.Marshal(file{Name: deck.Name, Cards: deck.List()})
	if err != nil {
		return fmt.Errorf("marshall deck: %w", err)
	}

	if err := os.WriteFile(deck.filename, data, 0644); err != nil {
		return fmt.Errorf("write deck: %w", err)
	}

	return nil
}

// Remove removes the deck from the repository.
func (r *Repository) Remove(deck *Deck) error {
	if deck == nil {
		return ErrDeckNotExist
	}

	if _, ok := r.decks[deck.id]; !ok {
		return ErrDeckNotExist
	}

	if err := os.Remove(deck.filename); err != nil {
		return fmt.Errorf("remove deck '%s': %w", deck.Name, err)
	}

	delete(r.decks, deck.id)

	return nil
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

func loadDecks(path string, clock Clock) (map[string]*Deck, error) {
	files, err := filepath.Glob(path + "/*.toml")
	if err != nil {
		return nil, fmt.Errorf("find decks: %w", err)
	}

	decks := make(map[string]*Deck, len(files))
	for _, file := range files {
		deck, err := OpenDeck(file, clock)
		if err != nil {
			return nil, err
		}
		decks[deck.id] = deck
	}
	return decks, nil
}
