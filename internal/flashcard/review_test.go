package flashcard_test

import (
	"testing"
	"time"

	"github.com/open-spaced-repetition/go-fsrs/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eliostvs/lembrol/internal/clock"
	testclock "github.com/eliostvs/lembrol/internal/clock/test"
	"github.com/eliostvs/lembrol/internal/flashcard"
)

func TestNewReview(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		clock clock.Clock
		deck  string
		want  int
	}{
		{
			name:  "when deck has no due cards",
			clock: testclock.New(beforeOldestCard),
			deck:  largeDeck,
			want:  0,
		},
		{
			name:  "when deck has due cards",
			clock: clock.New(),
			deck:  largeDeck,
			want:  7,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				review := newTestReview(t, tt.deck, tt.clock)

				assert.Equal(t, tt.want, review.Left())
			},
		)
	}
}

func TestReview_Rate(t *testing.T) {
	t.Parallel()

	t.Run(
		"returns error when queue is empty", func(t *testing.T) {
			review := newTestReview(t, smallDeck, testclock.New(beforeOldestCard))

			newReview, err := review.Rate(flashcard.ReviewScoreGood)

			assert.Equal(t, flashcard.Review{}, newReview)
			assert.ErrorIs(t, err, flashcard.ErrEmptyReview)
		},
	)

	t.Run(
		"advances card", func(t *testing.T) {
			type args struct {
				deck  string
				time  time.Time
				score flashcard.ReviewScore
			}
			type want struct {
				current, left, completed, total int
			}
			tests := []struct {
				name string
				args
				want
			}{
				{
					name: "re-queues the card when review score is again",
					args: args{
						deck:  largeDeck,
						time:  time.Now(),
						score: flashcard.ReviewScoreAgain,
					},
					want: want{
						total:     7,
						left:      7,
						current:   1,
						completed: 0,
					},
				},
				{
					name: "removes card from the queue",
					args: args{
						deck:  largeDeck,
						time:  time.Now(),
						score: flashcard.ReviewScoreGood,
					},
					want: want{
						total:     7,
						left:      6,
						current:   2,
						completed: 1,
					},
				},
				{
					name: "current is never greater than total",
					args: args{
						deck:  singleDeck,
						time:  time.Now(),
						score: flashcard.ReviewScoreGood,
					},
					want: want{
						total:     1,
						left:      0,
						current:   1,
						completed: 1,
					},
				},
			}
			for _, tt := range tests {
				t.Run(
					tt.name, func(t *testing.T) {
						review := newTestReview(t, tt.args.deck, testclock.New(tt.args.time))
						card, err := review.Card()
						require.NoError(t, err)

						assert.Contains(t, review.Deck.DueCards(), card)
						newReview, err := review.Rate(tt.args.score)

						newCard := getCard(newReview.Deck, card.ID)
						assert.Greater(t, len(newCard.Stats), len(card.Stats))

						assert.NoError(t, err)
						assert.Equal(t, tt.want.left, newReview.Left())
						assert.Equal(t, tt.want.current, newReview.Current())
						assert.Equal(t, tt.want.total, newReview.Total())
						assert.Equal(t, tt.want.completed, newReview.Completed)
					},
				)
			}
		},
	)
}

func TestReview_Skip(t *testing.T) {
	t.Parallel()

	t.Run(
		"returns error when queue is empty", func(t *testing.T) {
			review := newTestReview(t, largeDeck, testclock.New(beforeOldestCard))

			review, err := review.Skip()

			assert.Equal(t, flashcard.Review{}, review)
			assert.ErrorIs(t, err, flashcard.ErrEmptyReview)
		},
	)

	t.Run(
		"moves current card to the end of the queue", func(t *testing.T) {
			review := newTestReview(t, smallDeck, clock.New())
			card, _ := review.Card()

			review, err := review.Skip()
			require.NoError(t, err)

			nextCard, _ := review.Card()
			assert.NotEqual(t, card, nextCard)

			review, err = review.Skip()
			require.NoError(t, err)
			review, err = review.Skip()
			require.NoError(t, err)

			nextCard, _ = review.Card()
			assert.Equal(t, card, nextCard)
		},
	)
}

func TestReviewScoreToFSRSRating(t *testing.T) {
	tests := []struct {
		score    flashcard.ReviewScore
		expected fsrs.Rating
	}{
		{flashcard.ReviewScoreAgain, fsrs.Again},
		{flashcard.ReviewScoreHard, fsrs.Hard},
		{flashcard.ReviewScoreGood, fsrs.Good},
		{flashcard.ReviewScoreEasy, fsrs.Easy},
	}

	for _, tt := range tests {
		t.Run(tt.score.String(), func(t *testing.T) {
			result := flashcard.ReviewScoreToFSRSRating(tt.score)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFSRSRatingToReviewScore(t *testing.T) {
	tests := []struct {
		rating   fsrs.Rating
		expected flashcard.ReviewScore
	}{
		{fsrs.Again, flashcard.ReviewScoreAgain},
		{fsrs.Hard, flashcard.ReviewScoreHard},
		{fsrs.Good, flashcard.ReviewScoreGood},
		{fsrs.Easy, flashcard.ReviewScoreEasy},
	}

	for _, tt := range tests {
		t.Run(string(rune('0'+int(tt.rating))), func(t *testing.T) {
			result := flashcard.FSRSRatingToReviewScore(tt.rating)
			assert.Equal(t, tt.expected, result)
		})
	}
}

/*
 Test Utilities
*/

func getCard(deck flashcard.Deck, id string) flashcard.Card {
	for _, card := range deck.List() {
		if card.ID == id {
			return card
		}
	}

	return flashcard.Card{}
}

func newTestReview(t *testing.T, file string, c clock.Clock) flashcard.Review {
	t.Helper()

	return flashcard.NewReview(newTestDeck(t, file, c), c)
}
