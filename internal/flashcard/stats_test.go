package flashcard_test

import (
	"testing"
	"time"

	"github.com/eliostvs/lembrol/internal/flashcard"
	"github.com/open-spaced-repetition/go-fsrs/v3"
	"github.com/stretchr/testify/assert"
)

func TestNewStats(t *testing.T) {
	now := time.Now()
	previous := flashcard.Card{
		LastReview: now.Add(-time.Hour),
		Reps:       1,
	}
	got := flashcard.Card{
		Stability:     2.5,
		Difficulty:    3.2,
		ElapsedDays:   1,
		ScheduledDays: 3,
		State:         fsrs.Review,
	}

	stats := flashcard.NewStats(now, fsrs.Good, previous, got)

	assert.Equal(t, flashcard.ReviewScoreNormal, stats.Score)
	assert.Equal(t, fsrs.Good, stats.Rating)
	assert.Equal(t, now, stats.LastReview)
	assert.Equal(t, got.Stability, stats.Stability)
	assert.Equal(t, got.Difficulty, stats.Difficulty)
	assert.Equal(t, got.ElapsedDays, stats.ElapsedDays)
	assert.Equal(t, got.ScheduledDays, stats.ScheduledDays)
	assert.Equal(t, got.State, stats.State)
}
