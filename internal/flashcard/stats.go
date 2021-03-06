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
	Algorithm      string      `json:"algorithm"`
	CardID         string      `json:"card_id"`
	Timestamp      time.Time   `json:"timestamp"`
	Score          ReviewScore `json:"score,string"`
	LastReview     time.Time   `json:"last_review"`
	Repetitions    int         `json:"repetitions"`
	Interval       float64     `json:"interval,string"`
	EasinessFactor float64     `json:"easiness_factor,string"`
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

	file, err := os.OpenFile(r.path(deck.Name), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open stats %s: %w", r.path(deck.Name), err)
	}

	_, err = file.Write(append(data, '\n'))
	if err0 := file.Close(); err0 != nil && err == nil {
		err = err0
	}

	if err != nil {
		return fmt.Errorf("save stats: %w", err)
	}

	return nil
}

// Find returns the stats from a given card.
func (r StatsRepository) Find(deck Deck, card Card) ([]Stats, error) {
	file, err := os.Open(r.path(deck.Name))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open stats %s : %w", r.path(deck.Name), err)
	}

	found := make([]Stats, 0, 0)
	decoder := json.NewDecoder(file)
	for {
		var stats Stats

		if err := decoder.Decode(&stats); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("unmarshalling stats: %w", err)
		}

		if stats.CardID == card.ID() {
			found = append(found, stats)
		}
	}

	return found, nil
}

func (r StatsRepository) path(name string) string {
	return filepath.Join(r.location, slugify.Slugify(name)+"-stats.jsonl")
}
