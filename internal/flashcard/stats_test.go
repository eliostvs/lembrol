package flashcard_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eliostvs/remembercli/internal/flashcard"
	"github.com/eliostvs/remembercli/internal/test"
)

func TestStatsRepository_Save(t *testing.T) {
	t.Run("returns error when save stats fail", func(t *testing.T) {
		location, cleanup := test.TempReadOnlyDirCopy(t, t.TempDir())
		t.Cleanup(cleanup)
		deck := newDeck(t, smallDeck)
		repo := flashcard.NewStatsRepository(location)

		err := repo.Save(deck, &flashcard.Stats{})

		assert.Error(t, err)
	})

	t.Run("save stats to disk", func(t *testing.T) {
		location := t.TempDir()
		deck := newDeck(t, smallDeck)
		repo := flashcard.NewStatsRepository(location)

		stats := flashcard.Stats{
			Algorithm:      "Algorithm",
			Card:           "Card",
			Timestamp:      "Timestamp",
			Score:          "1",
			LastReview:     "LastReview",
			Repetitions:    1,
			Interval:       "Interval",
			EasinessFactor: "EasinessFactor",
		}
		err := repo.Save(deck, &stats)
		assert.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(location, "golang-small-stats.jsonl"))
		require.NoError(t, err)
		want := `{"algorithm":"Algorithm","card":"Card","timestamp":"Timestamp","score":"1","last_review":"LastReview","repetitions":1,"interval":"Interval","easiness_factor":"EasinessFactor"}`
		assert.JSONEq(t, want, string(content))
	})
}
