package flashcard

import (
	"time"

	fsrs "github.com/open-spaced-repetition/go-fsrs/v3"
)

// Stats is the revised card statistics.
type Stats struct {
	Score       ReviewScore `json:"score" validate:"required"`
	LastReview  time.Time   `json:"last_review" validate:"required"`
	
	// FSRS-specific fields 
	Rating        fsrs.Rating `json:"rating"`
	Stability     float64     `json:"stability"`
	Difficulty    float64     `json:"difficulty"`
	ElapsedDays   uint64      `json:"elapsed_days"`
	ScheduledDays uint64      `json:"scheduled_days"`
	State         fsrs.State  `json:"state"`
	
}


// NewFSRSStats creates stats using FSRS data.
func NewFSRSStats(ts time.Time, rating fsrs.Rating, previous, updated Card) Stats {
	return Stats{
		Score:         FSRSRatingToReviewScore(rating),
		Rating:        rating,
		LastReview:    ts,
		Stability:     updated.Stability,
		Difficulty:    updated.Difficulty,
		ElapsedDays:   updated.ElapsedDays,
		ScheduledDays: updated.ScheduledDays,
		State:         updated.State,
	}
}
