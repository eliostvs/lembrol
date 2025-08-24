package flashcard

import (
	"errors"
	"math/rand"
	"strconv"

	"github.com/eliostvs/lembrol/internal/clock"
	"github.com/open-spaced-repetition/go-fsrs/v3"
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

	if number < 1 || number > 4 {
		return ReviewScore(-1), ErrInvalidScore
	}

	return ReviewScore(number), nil
}

// ReviewScore defines grade for review attempts.
// Review uses scores to calculate rating in range from [1, 4] inclusive.
type ReviewScore int

func (s ReviewScore) String() string {
	return strconv.Itoa(int(s))
}

const (
	ReviewScoreAgain ReviewScore = iota + 1
	ReviewScoreHard
	ReviewScoreGood
	ReviewScoreEasy
)

var Scores = [4]ReviewScore{
	ReviewScoreAgain,
	ReviewScoreHard,
	ReviewScoreGood,
	ReviewScoreEasy,
}

// ErrEmptyReview indicates the review session has not more cards left to review.
var ErrEmptyReview = errors.New("no cards in queue")

// NewReview returns a new Review from a given a deck.
// It gets the due cards from the deck a shuffle them.
func NewReview(deck Deck, clock clock.Clock) Review {
	dueCards := deck.DueCards()
	shuffle(dueCards)
	return Review{queue: dueCards, Deck: deck, clock: clock, scheduler: DefaultScheduler()}
}

func shuffle(cards []Card) {
	rand.Shuffle(len(cards), func(i, j int) { cards[i], cards[j] = cards[j], cards[i] })
}

// Review represents a review session.
type Review struct {
	Deck      Deck
	queue     []Card
	clock     clock.Clock
	scheduler *Scheduler
	Completed int
}

// Total returns the number of cards in the review session.
func (r Review) Total() int {
	return r.Completed + r.Left()
}

// Left returns the number of cards left to review.
func (r Review) Left() int {
	return len(r.queue)
}

// Current returns the number of cards already reviewed.
func (r Review) Current() int {
	if r.Completed == r.Total() {
		return r.Completed
	}
	return r.Completed + 1
}

// Rate scores the current card.
func (r Review) Rate(score ReviewScore) (Review, error) {
	card, err := r.Card()
	if err != nil {
		return Review{}, err
	}

	rating := ReviewScoreToFSRSRating(score)
	ts := r.clock.Now()
	card = r.scheduler.ScheduleCard(card, ts, rating)

	r.queue = r.queue[1:]
	r.Deck = r.Deck.Change(card)

	// For "Again" ratings, add card back to queue without advancing
	if rating == fsrs.Again {
		r.queue = append(r.queue, card)
	} else {
		r.Completed++
	}

	return r, nil
}

// Skip moves the current card to the end of the queue.
func (r Review) Skip() (Review, error) {
	card, err := r.Card()
	if err != nil {
		return Review{}, err
	}

	r.queue = r.queue[1:]
	r.queue = append(r.queue, card)

	return r, nil
}

// Card returns the card being reviewed.
func (r Review) Card() (Card, error) {
	if len(r.queue) == 0 {
		return Card{}, ErrEmptyReview
	}
	return r.queue[0], nil
}
