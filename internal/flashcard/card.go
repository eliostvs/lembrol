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

// NewCard create a new Card instance.
func NewCard(question, answer string, today time.Time) Card {
	return Card{
		id:             nanoid.Must(),
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
	Interval       float64
	Repetitions    int

	id string
}

// Advance advances supermemo state for a card.
func (c Card) Advance(ts time.Time, score ReviewScore) (Card, *Stats) {
	previous := c
	c.ReviewedAt = ts

	if score < ReviewScoreNormal {
		c.Repetitions = 0
		c.Interval = 1
		return c, c.stats(ts, score, previous)
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

	return c, c.stats(ts, score, previous)
}

func (c Card) nextInterval() float64 {
	return math.Ceil(c.Interval * c.EasinessFactor)
}

func (c Card) nextEasinessFactor(score ReviewScore) float64 {
	newEasinessFactor := roundNearest(c.EasinessFactor + (0.1 - (5-float64(score))*(0.08+(5-float64(score))*0.02)))
	return math.Max(MinEasinessFactor, newEasinessFactor)
}

func (Card) stats(ts time.Time, score ReviewScore, previous Card) *Stats {
	return &Stats{
		Algorithm:      "sm2",
		Timestamp:      ts,
		CardID:         previous.id,
		Score:          score,
		LastReview:     previous.ReviewedAt,
		Repetitions:    previous.Repetitions,
		Interval:       previous.Interval,
		EasinessFactor: previous.EasinessFactor,
	}
}

// NextReviewAt returns next review timestamp for a card.
func (c Card) NextReviewAt() time.Time {
	return c.ReviewedAt.Add(time.Duration(hoursPerDay*c.Interval) * time.Hour)
}

// IsDue reports whether the card is due at the instant t.
func (c Card) IsDue(t time.Time) bool {
	return c.NextReviewAt().Before(t)
}

// ID returns card identifier.
func (c Card) ID() string {
	return c.id
}

func roundNearest(x float64) float64 {
	return math.Round(x*100) / 100
}
