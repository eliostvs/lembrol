package flashcard

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/avelino/slugify"
)

// Stats is the revised card statistics.
type Stats struct {
	Algorithm      string    `json:"algorithm"`
	CardID         string    `json:"card_id"`
	Timestamp      time.Time `json:"timestamp"`
	Score          int       `json:"score,string"`
	LastReview     time.Time `json:"last_review"`
	Repetitions    int       `json:"repetitions"`
	Interval       float64   `json:"interval,string"`
	EasinessFactor float64   `json:"easiness_factor,string"`
}

func NewStatsRepository(location string) StatsRepository {
	return StatsRepository{
		location: location,
	}
}

// StatsRepository defines the storage interface that manages stats.
type StatsRepository struct {
	location string
}

// Save writes stats to disk.
func (r StatsRepository) Save(deck Deck, stats *Stats) error {
	data, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("marshal stats: %w", err)
	}

	file, err := os.OpenFile(StatsPath(r.location, deck.Name), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open stats file: %w", err)
	}

	_, err = file.Write(append(data, '\n'))
	if err0 := file.Close(); err0 != nil && err == nil {
		err = err0
	}

	if err != nil {
		return fmt.Errorf("write stats: %w", err)
	}

	return nil
}

// Find returns the stats from a given card.
func (r StatsRepository) Find(deck Deck, card Card) ([]Stats, error) {
	file, err := os.Open(StatsPath(r.location, deck.Name))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open stats file: %w", err)
	}

	found := make([]Stats, 0, 0)
	decoder := json.NewDecoder(file)
	for {
		var stats Stats

		err := decoder.Decode(&stats)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("unmarshalling stats: %w", err)
		}

		if stats.CardID == card.ID() {
			found = append(found, stats)
		}
	}

	return found, nil
}

func StatsPath(location, deck string) string {
	return filepath.Join(location, slugify.Slugify(deck)+"-stats.jsonl")
}
