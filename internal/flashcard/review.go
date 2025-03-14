package flashcard

import (
	"errors"
	"math/rand"

	"github.com/eliostvs/lembrol/internal/clock"
)

// ErrEmptyReview indicates the review session has not more cards left to review.
var ErrEmptyReview = errors.New("no cards in queue")

// NewReview returns a new Review from a given a deck.
// It gets the due cards from the deck a shuffle them.
func NewReview(deck Deck, clock clock.Clock) Review {
	dueCards := deck.DueCards()
	shuffle(dueCards)
	return Review{queue: dueCards, Deck: deck, clock: clock}
}

func shuffle(cards []Card) {
	rand.Shuffle(len(cards), func(i, j int) { cards[i], cards[j] = cards[j], cards[i] })
}

// Review represents a review session.
type Review struct {
	Deck      Deck
	queue     []Card
	clock     clock.Clock
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

	if score == ReviewScoreAgain {
		r.queue = r.queue[1:]
		r.queue = append(r.queue, card)
		return r, nil
	}

	card = card.Advance(r.clock.Now(), score)
	r.queue = r.queue[1:]
	r.Completed++
	r.Deck = r.Deck.Change(card)
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
