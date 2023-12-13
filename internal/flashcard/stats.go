package flashcard

import (
	"time"
)

// Stats is the revised card statistics.
type Stats struct {
	Algorithm      string      `json:"algorithm" validate:"required"`
	Timestamp      time.Time   `json:"timestamp" validate:"required"`
	Score          ReviewScore `json:"score" validate:"required"`
	LastReview     time.Time   `json:"last_review" validate:"required"`
	Repetitions    int         `json:"repetitions" validate:"required"`
	Interval       float64     `json:"interval" validate:"required"`
	EasinessFactor float64     `json:"easiness_factor" validate:"required"`
}

func NewStats(ts time.Time, score ReviewScore, previous Card) Stats {
	return Stats{
		Algorithm:      "sm2",
		Timestamp:      ts,
		Score:          score,
		LastReview:     previous.ReviewedAt,
		Repetitions:    previous.Repetitions,
		Interval:       previous.Interval,
		EasinessFactor: previous.EasinessFactor,
	}
}
