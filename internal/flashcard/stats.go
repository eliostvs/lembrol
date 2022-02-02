package flashcard

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/avelino/slugify"
)

func NewStatsRepository(location string) *StatsRepository {
	return &StatsRepository{
		location: location,
	}
}

// StatsRepository defines the storage interface that manages stats.
type StatsRepository struct {
	location string
}

// Save writes stats to disk.
func (s *StatsRepository) Save(deck Deck, stats *Stats) error {
	data, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("marshal stats: %w", err)
	}

	f, err := os.OpenFile(s.path(deck), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open stats file: %w", err)
	}

	_, err = f.Write(append(data, '\n'))
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}
	if err != nil {
		return fmt.Errorf("write stats: %w", err)
	}

	return nil
}

func (s *StatsRepository) path(d Deck) string {
	return filepath.Join(s.location, slugify.Slugify(d.Name)+"-stats.jsonl")
}
