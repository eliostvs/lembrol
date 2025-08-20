package flashcard

import (
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"time"

	nanoid "github.com/matoous/go-nanoid/v2"
	fsrs "github.com/open-spaced-repetition/go-fsrs/v3"
)

// ErrInvalidScore is returned by NewReviewScore
// when the given string is not a number of is out of the valid range.
var ErrInvalidScore = errors.New("invalid score")

// NewReviewScore converts a string to a ReviewScore.
func NewReviewScore(s string) (ReviewScore, error) {
	number, err := strconv.Atoi(s)
	if err != nil {
		return ReviewScore(-1), ErrInvalidScore
	}

	if number < 0 || number > 4 {
		return ReviewScore(-1), ErrInvalidScore
	}

	return ReviewScore(number), nil
}

// ReviewScore defines grade for review attempts.
// Review uses scores to calculate rating in range from [0, 4] inclusive.
type ReviewScore int

func (s ReviewScore) String() string {
	return strconv.Itoa(int(s))
}

const (
	ReviewScoreAgain ReviewScore = iota
	ReviewScoreHard
	ReviewScoreNormal
	ReviewScoreEasy
	ReviewScoreSuperEasy
)

var Scores = [5]ReviewScore{
	ReviewScoreAgain,
	ReviewScoreHard,
	ReviewScoreNormal,
	ReviewScoreEasy,
	ReviewScoreSuperEasy,
}

type cardOption func(Card) Card

func withStats(stats []Stats) cardOption {
	return func(card Card) Card {
		card.Stats = stats
		return card
	}
}

// NewCard create a new Card instance using FSRS.
func NewCard(question, answer string, today time.Time, options ...cardOption) Card {
	// Create base FSRS card
	fsrsCard := fsrs.NewCard()

	card := Card{
		ID:       nanoid.Must(),
		Question: question,
		Answer:   answer,

		// Initialize with FSRS defaults
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

	for _, fn := range options {
		card = fn(card)
	}

	return card
}

// Card represents a single card in a Deck.
type Card struct {
	ID       string    `json:"id" validate:"required"`
	Question string    `json:"question" validate:"required"`
	Answer   string    `json:"answer" validate:"required"`
	Stats    []Stats   `json:"stats"`

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

// Advance advances card state using FSRS algorithm.
func (c Card) Advance(ts time.Time, score ReviewScore) Card {
	return c.AdvanceWithScheduler(ts, score, DefaultScheduler())
}

// AdvanceWithScheduler advances card state using the provided FSRS scheduler.
func (c Card) AdvanceWithScheduler(ts time.Time, score ReviewScore, scheduler *Scheduler) Card {
	// Migrate card if it hasn't been migrated yet
	if !IsMigrated(c) {
		c = MigrateCard(c, ts)
	}

	// Convert ReviewScore to FSRS Rating
	rating := ReviewScoreToFSRSRating(score)

	// Use FSRS scheduler to get next state
	updatedCard, stats := scheduler.ScheduleCard(c, ts, rating)

	// Append stats to history
	updatedCard.Stats = append(c.Stats, stats)

	return updatedCard
}

// NextReviewAt returns next review timestamp for a card.
func (c Card) NextReviewAt() time.Time {
	// Use FSRS Due field
	if !c.Due.IsZero() {
		return c.Due
	}

	// Default to current time for new cards
	return c.LastReview
}

// IsDue reports whether the card is due at the instant t.
func (c Card) IsDue(t time.Time) bool {
	// Use FSRS Due field
	if !c.Due.IsZero() {
		return !c.Due.After(t)
	}

	// Default behavior for cards without due date
	return c.NextReviewAt().Before(t) || c.NextReviewAt().Equal(t)
}

func roundNearest(x float64) float64 {
	return math.Round(x*100) / 100
}

// UnmarshalJSON handles backward compatibility for the ReviewedAt field
func (c *Card) UnmarshalJSON(data []byte) error {
	// Create a temporary struct to unmarshal into
	type TempCard struct {
		ID         string    `json:"id"`
		Question   string    `json:"question"`
		Answer     string    `json:"answer"`
		Stats      []Stats   `json:"stats"`

		// FSRS-specific fields
		Due           time.Time  `json:"due"`
		Stability     float64    `json:"stability"`
		Difficulty    float64    `json:"difficulty"`
		ElapsedDays   uint64     `json:"elapsed_days"`
		ScheduledDays uint64     `json:"scheduled_days"`
		Reps          uint64     `json:"reps"`
		Lapses        uint64     `json:"lapses"`
		State         fsrs.State `json:"state"`
		LastReview    time.Time  `json:"last_review"`
		
		// Legacy field for backward compatibility
		ReviewedAt *time.Time `json:"reviewed_at"`
	}

	var temp TempCard
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Copy all fields
	c.ID = temp.ID
	c.Question = temp.Question
	c.Answer = temp.Answer
	c.Stats = temp.Stats
	c.Due = temp.Due
	c.Stability = temp.Stability
	c.Difficulty = temp.Difficulty
	c.ElapsedDays = temp.ElapsedDays
	c.ScheduledDays = temp.ScheduledDays
	c.Reps = temp.Reps
	c.Lapses = temp.Lapses
	c.State = temp.State

	// Handle backward compatibility for ReviewedAt -> LastReview
	if !temp.LastReview.IsZero() {
		c.LastReview = temp.LastReview
	} else if temp.ReviewedAt != nil && !temp.ReviewedAt.IsZero() {
		c.LastReview = *temp.ReviewedAt
	}

	return nil
}
