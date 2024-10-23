package flashcard

import (
	"errors"
	"math"
	"strconv"
	"time"

	nanoid "github.com/matoous/go-nanoid/v2"
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

	// InitialEasinessFactor defines the initial easiness factor.
	InitialEasinessFactor = 2.5
	// MinEasinessFactor defines the minimal easiness factor possible.
	MinEasinessFactor = 1.3
	hoursPerDay       = 24
)

var Scores = [5]ReviewScore{
	ReviewScoreAgain,
	ReviewScoreHard,
	ReviewScoreNormal,
	ReviewScoreEasy,
	ReviewScoreSuperEasy,
}

type cardOption func(Card) Card

func withInterval(interval float64) cardOption {
	return func(card Card) Card {
		card.Interval = interval
		return card
	}
}

func withRepetitions(repetitions int) cardOption {
	return func(card Card) Card {
		card.Repetitions = repetitions
		return card
	}
}

func withStats(stats []Stats) cardOption {
	return func(card Card) Card {
		card.Stats = stats
		return card
	}
}

// NewCard create a new Card instance.
func NewCard(question, answer string, today time.Time, options ...cardOption) Card {
	card := Card{
		ID:             nanoid.Must(),
		Question:       question,
		Answer:         answer,
		ReviewedAt:     today,
		EasinessFactor: InitialEasinessFactor,
	}

	for _, fn := range options {
		card = fn(card)
	}

	return card
}

// Card represents a single card in a Deck.
type Card struct {
	ID             string    `json:"id" validate:"required"`
	Question       string    `json:"question" validate:"required"`
	Answer         string    `json:"answer" validate:"required"`
	ReviewedAt     time.Time `json:"reviewed_at" validate:"required"`
	EasinessFactor float64   `json:"easiness_factor" validate:"required"`
	Interval       float64   `json:"interval" validate:"required"`
	Repetitions    int       `json:"repetitions" validate:"required"`
	Stats          []Stats   `json:"stats"`
}

// Advance advances supermemo state for a card.
func (c Card) Advance(ts time.Time, score ReviewScore) Card {
	previous := c
	c.ReviewedAt = ts

	if score < ReviewScoreNormal {
		c.Repetitions = 0
		c.Interval = 1
		c.Stats = append(c.Stats, NewStats(ts, score, previous))
		return c
	}

	switch c.Repetitions {
	case 0:
		c.Interval = 1
	case 1:
		c.Interval = 6
	default:
		c.Interval = c.nextInterval()
	}
	c.Repetitions++
	c.EasinessFactor = c.nextEasinessFactor(score)
	c.Stats = append(c.Stats, NewStats(ts, score, previous))

	return c
}

func (c Card) nextInterval() float64 {
	return math.Ceil(c.Interval * c.EasinessFactor)
}

func (c Card) nextEasinessFactor(score ReviewScore) float64 {
	newEasinessFactor := roundNearest(c.EasinessFactor + (0.1 - (5-float64(score))*(0.08+(5-float64(score))*0.02)))
	return math.Max(MinEasinessFactor, newEasinessFactor)
}

// NextReviewAt returns next review timestamp for a card.
func (c Card) NextReviewAt() time.Time {
	return c.ReviewedAt.Add(time.Duration(hoursPerDay*c.Interval) * time.Hour)
}

// IsDue reports whether the card is due at the instant t.
func (c Card) IsDue(t time.Time) bool {
	return c.NextReviewAt().Before(t)
}

func roundNearest(x float64) float64 {
	return math.Round(x*100) / 100
}
