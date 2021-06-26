package flashcard

import (
	"errors"
	"math/rand"
	"time"
)

// ErrEmptyReview indicates the review session has not more cards left to review.
var ErrEmptyReview = errors.New("no cards in queue")

// NewReview returns a new Review from a given a deck.
// It get the due cards from the deck an shuffle them.
func NewReview(deck Deck, clock Clock) *Review {
	dueCards := deck.DueCards()
	shuffle(dueCards)
	return &Review{queue: dueCards, deck: deck, clock: clock}
}

func shuffle(cards []Card) {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(cards), func(i, j int) { cards[i], cards[j] = cards[j], cards[i] })
}

// Review represents a review session.
type Review struct {
	deck      Deck
	queue     []Card
	clock     Clock
	completed int
}

// Total returns the number of cards in the review session.
func (r *Review) Total() int {
	return r.completed + r.Left()
}

// Left returns the number of cards left to review.
func (r *Review) Left() int {
	return len(r.queue)
}

// Current returns the number of cards already reviewed.
func (r *Review) Current() int {
	if r.completed == r.Total() {
		return r.Total()
	}
	return r.completed + 1
}

// Completed returns the number of cards already reviewed.
func (r *Review) Completed() int {
	return r.completed
}

func (r *Review) Rate(score ReviewScore) (Card, error) {
	card, err := r.CurrentCard()
	if err != nil {
		return card, err
	}

	if score == ReviewScoreAgain {
		r.queue = r.queue[1:]
		r.queue = append(r.queue, card)
		return card, nil
	}

	card = card.Advance(r.clock.Now(), score)
	r.queue = r.queue[1:]
	r.completed++
	r.deck = r.deck.Change(card)
	return card, nil
}

// Skip moves the current card to the end of the queue.
func (r *Review) Skip() error {
	card, err := r.CurrentCard()
	if err != nil {
		return err
	}

	r.queue = r.queue[1:]
	r.queue = append(r.queue, card)

	return nil
}

// CurrentCard returns the card being reviewed.
func (r *Review) CurrentCard() (Card, error) {
	if len(r.queue) == 0 {
		return Card{}, ErrEmptyReview
	}
	return r.queue[0], nil
}

// Deck returns the deck being reviewed.
func (r *Review) Deck() Deck {
	return r.deck
}
