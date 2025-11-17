package flashcard

import (
	"time"

	fsrs "github.com/open-spaced-repetition/go-fsrs/v3"
)

// Stats is the revised card statistics.
type Stats struct {
	Rating        fsrs.Rating `json:"rating"`
	Stability     float64     `json:"stability"`
	Difficulty    float64     `json:"difficulty"`
	ElapsedDays   uint64      `json:"elapsed_days"`
	ScheduledDays uint64      `json:"scheduled_days"`
	Reps          uint64      `json:"reps"`
	Lapses        uint64      `json:"lapses"`
	State         fsrs.State  `json:"state"`
	LastReview    time.Time   `json:"last_review"`
}

// NewStats creates stats using FSRS data.
func NewStats(ts time.Time, rating fsrs.Rating, previous, updated Card) Stats {
	return Stats{
		Rating:        rating,
		Stability:     updated.Stability,
		Difficulty:    updated.Difficulty,
		ElapsedDays:   updated.ElapsedDays,
		ScheduledDays: updated.ScheduledDays,
		Reps:          updated.Reps,
		Lapses:        updated.Lapses,
		State:         updated.State,
		LastReview:    ts,
	}
}
