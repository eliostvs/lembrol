package flashcard_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eliostvs/lembrol/internal/flashcard"
	"github.com/eliostvs/lembrol/internal/test"
)

func TestNewReview(t *testing.T) {
	tests := []struct {
		name  string
		clock flashcard.Clock
		deck  string
		want  int
	}{
		{
			name:  "when deck has no due cards",
			clock: test.NewClock(beforeOldestCard),
			deck:  largeDeck,
			want:  0,
		},
		{
			name:  "when deck has due cards",
			clock: flashcard.NewClock(),
			deck:  largeDeck,
			want:  7,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			review := flashcard.NewReview(newDeck(t, tt.deck, func(o *option) {
				o.clock = tt.clock
			}), tt.clock)

			assert.Equal(t, tt.want, review.Left())
			assert.Equal(t, tt.deck, review.Deck().Name)
		})
	}
}

func TestReview_Rate(t *testing.T) {
	t.Run("returns error when queue is empty", func(t *testing.T) {
		review := newReview(t, smallDeck, withTestClock(beforeOldestCard))

		stats, review, err := review.Rate(flashcard.ReviewScoreNormal)

		assert.Nil(t, stats)
		assert.Equal(t, flashcard.Review{}, review)
		assert.ErrorIs(t, err, flashcard.ErrEmptyReview)
	})

	t.Run("advances card", func(t *testing.T) {
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
					score: flashcard.ReviewScoreNormal,
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
					deck:  singleCardDeck,
					time:  time.Now(),
					score: flashcard.ReviewScoreNormal,
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
			t.Run(tt.name, func(t *testing.T) {
				review := newReview(t, tt.args.deck, withTestClock(tt.args.time))
				card, _ := review.CurrentCard()

				assert.Contains(t, review.Deck().DueCards(), card)
				stats, newReview, err := review.Rate(tt.args.score)

				if tt.args.score == flashcard.ReviewScoreAgain {
					assert.Nil(t, stats)
				} else {
					assert.NotNil(t, stats)
				}

				assert.NoError(t, err)
				assert.Equal(t, tt.want.left, newReview.Left())
				assert.Equal(t, tt.want.current, newReview.Current())
				assert.Equal(t, tt.want.total, newReview.Total())
				assert.Equal(t, tt.want.completed, newReview.Completed())
			})
		}
	})
}

func TestReview_Skip(t *testing.T) {
	t.Run("returns error when queue is empty", func(t *testing.T) {
		review := newReview(t, largeDeck, withTestClock(beforeOldestCard))

		review, err := review.Skip()

		assert.Equal(t, flashcard.Review{}, review)
		assert.ErrorIs(t, err, flashcard.ErrEmptyReview)
	})

	t.Run("moves current card to the end of the queue", func(t *testing.T) {
		review := newReview(t, smallDeck)
		card, _ := review.CurrentCard()

		review, err := review.Skip()
		require.NoError(t, err)

		nextCard, _ := review.CurrentCard()
		assert.NotEqual(t, card, nextCard)

		review, err = review.Skip()
		require.NoError(t, err)
		review, err = review.Skip()
		require.NoError(t, err)

		nextCard, _ = review.CurrentCard()
		assert.Equal(t, card, nextCard)
	})
}

/*
 Test Utilities
*/

func newReview(t *testing.T, deck string, configs ...configOption) flashcard.Review {
	t.Helper()

	opts := option{clock: flashcard.NewClock()}
	for _, config := range configs {
		config(&opts)
	}

	return flashcard.NewReview(newDeck(t, deck, configs...), opts.clock)
}
