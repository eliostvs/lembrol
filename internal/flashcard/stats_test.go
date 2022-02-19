package flashcard_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eliostvs/lembrol/internal/flashcard"
	"github.com/eliostvs/lembrol/internal/test"
)

func TestStatsRepository_Save(t *testing.T) {
	t.Run("returns error when save stats fail", func(t *testing.T) {
		location, cleanup := test.TempReadOnlyDirCopy(t, manyDecksLocation)
		t.Cleanup(cleanup)
		deck := newDeck(t, largeDeck, withLocation(location))
		repo := flashcard.NewStatsRepository(location)

		err := repo.Save(deck, &flashcard.Stats{})

		assert.Error(t, err)
	})

	t.Run("save stats to disk", func(t *testing.T) {
		location := test.TempDirCopy(t, emptyDeckLocation)
		repo := flashcard.NewStatsRepository(location)
		deck := newDeck(t, emptyDeck, withLocation(location))

		stats := flashcard.Stats{
			Algorithm:      "sm2",
			CardID:         "99",
			Timestamp:      time.Date(2021, 01, 02, 00, 00, 00, 0, time.UTC),
			Score:          flashcard.ReviewScoreHard,
			LastReview:     time.Date(2021, 01, 02, 00, 00, 00, 0, time.UTC),
			Repetitions:    1,
			Interval:       10,
			EasinessFactor: 1.75,
		}
		err := repo.Save(deck, &stats)
		assert.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(location, "empty-stats.jsonl"))
		require.NoError(t, err)
		want := `{"algorithm":"sm2","card_id":"99","timestamp":"2021-01-02T00:00:00Z","score":"1","last_review":"2021-01-02T00:00:00Z","repetitions":1,"interval":"10","easiness_factor":"1.75"}`
		assert.JSONEq(t, want, string(content))
	})
}

func TestStatsRepository_Find(t *testing.T) {
	t.Run("returns no stats when file doesn't exist", func(t *testing.T) {
		location := test.TempDirCopy(t, manyDecksLocation)
		repo := flashcard.NewStatsRepository(location)
		deck := newDeck(t, singleCardDeck, withLocation(location))

		for _, card := range deck.List() {
			t.Run(card.Question, func(t *testing.T) {
				stats, err := repo.Find(deck, card)

				assert.NoError(t, err)
				assert.Empty(t, stats)
			})
		}
	})

	t.Run("returns no stats when the card does not have stats", func(t *testing.T) {
		location := test.TempDirCopy(t, manyDecksLocation)
		repo := flashcard.NewStatsRepository(location)
		deck := newDeck(t, smallDeck, withLocation(location))

		for _, card := range deck.List() {
			t.Run(card.Question, func(t *testing.T) {
				stats, err := repo.Find(deck, card)

				assert.NoError(t, err)
				assert.Empty(t, stats)
			})
		}
	})

	t.Run("returns error when stats file is invalid", func(t *testing.T) {
		location := test.TempDirCopy(t, "./testdata/invalid-stats")
		repo := flashcard.NewStatsRepository(location)
		deck := newDeck(t, "Invalid", withLocation(location))

		for _, card := range deck.List() {
			t.Run(card.Question, func(t *testing.T) {
				stats, err := repo.Find(deck, card)

				assert.Error(t, err)
				assert.Nil(t, stats)
			})
		}
	})

	t.Run("returns cards stats", func(t *testing.T) {
		location := test.TempDirCopy(t, manyDecksLocation)
		repo := flashcard.NewStatsRepository(location)
		deck := newDeck(t, largeDeck, withLocation(location))
		wantStats := map[string][]flashcard.Stats{
			"1": {{
				Algorithm:      "sm2",
				CardID:         "1",
				Timestamp:      time.Date(2022, 02, 03, 17, 00, 00, 0, time.UTC),
				Score:          flashcard.ReviewScoreNormal,
				LastReview:     time.Date(2021, 01, 02, 15, 00, 00, 0, time.UTC),
				Repetitions:    0,
				Interval:       0,
				EasinessFactor: 2.5,
			}},
			"2": {{
				Algorithm:      "sm2",
				CardID:         "2",
				Timestamp:      time.Date(2022, 02, 03, 17, 00, 00, 0, time.UTC),
				Score:          flashcard.ReviewScoreNormal,
				LastReview:     time.Date(2021, 01, 02, 15, 00, 00, 0, time.UTC),
				Repetitions:    0,
				Interval:       0,
				EasinessFactor: 2.5,
			}},
			"3": {{
				Algorithm:      "sm2",
				CardID:         "3",
				Timestamp:      time.Date(2022, 02, 03, 17, 00, 00, 0, time.UTC),
				Score:          flashcard.ReviewScoreSuperEasy,
				LastReview:     time.Date(2021, 01, 02, 15, 00, 00, 0, time.UTC),
				Repetitions:    0,
				Interval:       0,
				EasinessFactor: 2.5,
			}},
			"4": {
				{
					Algorithm:      "sm2",
					CardID:         "4",
					Timestamp:      time.Date(2022, 02, 03, 17, 00, 00, 0, time.UTC),
					Score:          flashcard.ReviewScoreEasy,
					LastReview:     time.Date(2021, 01, 02, 15, 00, 00, 0, time.UTC),
					Repetitions:    0,
					Interval:       0,
					EasinessFactor: 2.5,
				},
				{
					Algorithm:      "sm2",
					CardID:         "4",
					Timestamp:      time.Date(2022, 02, 03, 17, 00, 00, 0, time.UTC),
					Score:          flashcard.ReviewScoreHard,
					LastReview:     time.Date(2021, 01, 02, 15, 00, 00, 0, time.UTC),
					Repetitions:    0,
					Interval:       0,
					EasinessFactor: 2.5,
				},
			},
			"5": {{
				Algorithm:      "sm2",
				CardID:         "5",
				Timestamp:      time.Date(2022, 02, 03, 17, 00, 00, 0, time.UTC),
				Score:          flashcard.ReviewScoreEasy,
				LastReview:     time.Date(2021, 01, 02, 15, 00, 00, 0, time.UTC),
				Repetitions:    0,
				Interval:       0,
				EasinessFactor: 2.5,
			}},
			"6": {{
				Algorithm:      "sm2",
				CardID:         "6",
				Timestamp:      time.Date(2022, 02, 03, 17, 00, 00, 0, time.UTC),
				Score:          flashcard.ReviewScoreHard,
				LastReview:     time.Date(2021, 01, 02, 15, 00, 00, 0, time.UTC),
				Repetitions:    0,
				Interval:       0,
				EasinessFactor: 2.5,
			}},
			"7": {{
				Algorithm:      "sm2",
				CardID:         "7",
				Timestamp:      time.Date(2022, 02, 03, 17, 00, 00, 0, time.UTC),
				Score:          flashcard.ReviewScoreHard,
				LastReview:     time.Date(2021, 01, 02, 15, 00, 00, 0, time.UTC),
				Repetitions:    1,
				Interval:       0,
				EasinessFactor: 1.75,
			}},
		}

		for _, card := range deck.List() {
			t.Run(card.Question, func(t *testing.T) {
				stats, err := repo.Find(deck, card)

				assert.NoError(t, err)
				assert.ElementsMatch(t, stats, wantStats[card.ID()])
			})
		}
	})
}
