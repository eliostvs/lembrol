package flashcard

import (
	"encoding/csv"
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

// NewRepository create a new Repository by reading all decks
// from a given folder.
func NewRepository(directory string, clock Clock) (*Repository, error) {
	if err := assureDirExist(directory); err != nil {
		return nil, err
	}

	decks, err := loadDecks(directory, clock)
	if err != nil {
		return nil, err
	}
	return &Repository{decks: decks, clock: clock, directory: directory}, nil
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

// Repository defines storage interface that manages a set of decks.
type Repository struct {
	directory string
	decks     map[string]Deck
	clock     Clock
}

// List returns the available deck names.
func (r *Repository) List() []Deck {
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
func (r *Repository) Total() int {
	return len(r.decks)
}

// Create creates a new deck from a given name.
func (r *Repository) Create(name string) (Deck, error) {
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

func (r *Repository) deckPath(name string) string {
	return filepath.Join(r.directory, slugify.Slugify(name)+".toml")
}

// Open returns a deck given a name.
func (r Repository) Open(name string) (Deck, error) {
	for _, deck := range r.decks {
		if deck.Name == name {
			return deck, nil
		}
	}
	return Deck{}, ErrDeckNotExist
}

// Save writes changes to disk.
func (r *Repository) Save(deck Deck) error {
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
func (r *Repository) Remove(deck Deck) error {
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
func (r *Repository) SaveStats(stats Stats) error {
	f, err := os.OpenFile(r.statsPath(stats.Algorithm), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	w := csv.NewWriter(f)
	return w.WriteAll([][]string{stats.Data})
}

func (r *Repository) statsPath(name string) string {
	return filepath.Join(r.directory, name+".csv")
}
