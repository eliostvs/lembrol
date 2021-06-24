package flashcard

import (
	"errors"
	"math"
	"strconv"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
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
	// MinimalEasinessFactor defines the minimal easiness factor possible.
	MinimalEasinessFactor = 1.3
	hoursPerDay           = 24
)

// NewCard create a new Card instance.
func NewCard(question, answer string, today time.Time) *Card {
	return &Card{
		id:             gonanoid.Must(),
		Question:       question,
		Answer:         answer,
		ReviewedAt:     today,
		EasinessFactor: InitialEasinessFactor,
	}
}

// Card represents a single card in a Deck.
type Card struct {
	Question       string
	Answer         string
	ReviewedAt     time.Time
	EasinessFactor float64
	Interval       int
	Repetitions    int

	id string
}

// Advance advances supermemo state for a card.
func (c *Card) Advance(now time.Time, score ReviewScore) Card {
	c.ReviewedAt = now

	if score < ReviewScoreNormal {
		c.Repetitions = 0
		c.Interval = 1
		return *c
	}

	switch c.Repetitions {
	case 0:
		c.Interval = 1
	case 1:
		c.Interval = 6
	default:
		c.Interval = calcNextInterval(c.Interval, c.EasinessFactor)
	}
	c.Repetitions++
	c.EasinessFactor = calcNextEasinessFactor(c.EasinessFactor, MinimalEasinessFactor, score)

	return *c
}

func calcNextInterval(interval int, easinessFactor float64) int {
	return int(math.Ceil(float64(interval) * easinessFactor))
}

// nolint:gomnd
func calcNextEasinessFactor(easinessFactor, minEasinessFactor float64, score ReviewScore) float64 {
	newEasinessFactor := roundNearest(easinessFactor + (0.1 - (5-float64(score))*(0.08+(5-float64(score))*0.02)))
	return math.Max(minEasinessFactor, newEasinessFactor)
}

// nolint:gomnd
func roundNearest(x float64) float64 {
	return math.Round(x*100) / 100
}

// NextReviewAt returns next review timestamp for a card.
func (c *Card) NextReviewAt() time.Time {
	return c.ReviewedAt.Add(time.Duration(hoursPerDay*c.Interval) * time.Hour)
}

// Due reports whether the card is due at the instant t.
func (c *Card) Due(t time.Time) bool {
	return c.NextReviewAt().Before(t)
}

// Id returns card identifier.
func (c *Card) Id() string {
	return c.id
}
