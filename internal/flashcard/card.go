package flashcard

import (
	"time"

	nanoid "github.com/matoous/go-nanoid/v2"
	"github.com/open-spaced-repetition/go-fsrs/v3"
)

// NewCard create a new Card instance.
func NewCard(question, answer string, today time.Time) Card {
	fsrsCard := fsrs.NewCard()

	card := Card{
		ID:       nanoid.Must(),
		Question: question,
		Answer:   answer,
		// FSRS defaults
		Due:           today,
		Stability:     fsrsCard.Stability,
		Difficulty:    fsrsCard.Difficulty,
		ElapsedDays:   fsrsCard.ElapsedDays,
		ScheduledDays: fsrsCard.ScheduledDays,
		Reps:          fsrsCard.Reps,
		Lapses:        fsrsCard.Lapses,
		State:         fsrsCard.State,
		LastReview:    today,
	}

	return card
}

// Card represents a single card in a Deck.
type Card struct {
	ID       string  `json:"id" validate:"required"`
	Question string  `json:"question" validate:"required"`
	Answer   string  `json:"answer" validate:"required"`
	Stats    []Stats `json:"stats"`
	// FSRS-specific fields
	Due           time.Time  `json:"due"`
	Stability     float64    `json:"stability"`
	Difficulty    float64    `json:"difficulty"`
	ElapsedDays   uint64     `json:"elapsed_days"`
	ScheduledDays uint64     `json:"scheduled_days"`
	Reps          uint64     `json:"reps"`
	Lapses        uint64     `json:"lapses"`
	State         fsrs.State `json:"state"`
	LastReview    time.Time  `json:"last_review,omitempty" validate:"required"`
}

func (c Card) AddStats(s Stats) Card {
	c.Stats = append(c.Stats, s)
	return c
}

// IsDue reports whether the card is due at the instant t.
func (c Card) IsDue(t time.Time) bool {
	if !c.Due.IsZero() {
		return !c.Due.After(t)
	}

	return c.LastReview.Before(t) || c.LastReview.Equal(t)
}
