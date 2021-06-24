package flashcard_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eliostvs/remembercli/internal/flashcard"
	"github.com/eliostvs/remembercli/internal/test"
)

func TestNewReview(t *testing.T) {
	tests := []struct {
		name  string
		clock flashcard.Clock
		want  int
	}{
		{
			name:  "when deck has no due cards",
			clock: test.NewClock(beforeOldestCard),
			want:  0,
		},
		{
			name:  "when deck has due cards",
			clock: flashcard.NewClock(),
			want:  7,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			review := flashcard.NewReview(newDeck(t, largeDeck, func(o *option) {
				o.clock = tt.clock
			}), tt.clock)

			assert.Equal(t, tt.want, review.Left())
		})
	}
}

func TestReview_CurrentCard(t *testing.T) {
	t.Run("returns card when session has cards", func(t *testing.T) {
		review := newReview(t, largeDeck)
		card, err := review.CurrentCard()

		assert.NoError(t, err)
		assert.NotNil(t, card)
	})

	t.Run("returns error when session has not due cards", func(t *testing.T) {
		review := newReview(t, largeDeck, withTestClock(beforeOldestCard))
		card, err := review.CurrentCard()

		assert.Nil(t, card)
		assert.Error(t, err)
	})
}

func TestReview_Rate(t *testing.T) {
	t.Run("return error when queue is empty", func(t *testing.T) {
		review := newReview(t, smallDeck, withTestClock(beforeOldestCard))

		card, err := review.Rate(flashcard.ReviewScoreNormal)

		assert.Error(t, err)
		assert.Nil(t, card)
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
			}, {
				name: "current is never greater than total",
				args: args{
					deck:  "Golang One",
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

				card, err := review.Rate(tt.args.score)
				assert.NoError(t, err)

				assert.NotNil(t, card)
				assert.Equal(t, tt.want.left, review.Left())
				assert.Equal(t, tt.want.current, review.Current())
				assert.Equal(t, tt.want.total, review.Total())
				assert.Equal(t, tt.want.completed, review.Completed())
			})
		}
	})

	t.Run("advances calculates the card next review date", func(t *testing.T) {
		review := newReview(t, smallDeck, withTestClock(time.Now()))

		card, err := review.Rate(flashcard.ReviewScoreHard)
		assert.NoError(t, err)

		assert.True(t, card.ReviewedAt.After(oldestCard.ReviewedAt))
	})
}

func TestReview_Deck(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{
			name: "small deck",
			args: smallDeck,
			want: "Golang Small",
		},
		{
			name: "large deck",
			args: largeDeck,
			want: "Golang Large",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			review := newReview(t, tt.args)

			got := review.Deck()

			require.NotNil(t, got)
			assert.Equal(t, tt.want, got.Name)
		})
	}
}

func newReview(t *testing.T, deckName string, cfgOpts ...configOption) *flashcard.Review {
	t.Helper()

	opts := option{clock: flashcard.NewClock()}
	for _, cfg := range cfgOpts {
		cfg(&opts)
	}

	return flashcard.NewReview(newDeck(t, deckName, cfgOpts...), opts.clock)
}

func TestReview_Skip(t *testing.T) {
	t.Run("returns error when session has not due cards", func(t *testing.T) {
		review := newReview(t, largeDeck, withTestClock(beforeOldestCard))
		err := review.Skip()

		assert.ErrorIs(t, err, flashcard.ErrEmptyReview)
	})

	t.Run("moves current card to the end of the queue", func(t *testing.T) {
		review := newReview(t, smallDeck)
		card, _ := review.CurrentCard()

		assert.NoError(t, review.Skip())
		nextCard, _ := review.CurrentCard()
		assert.NotEqual(t, card, nextCard)

		assert.NoError(t, review.Skip())
		assert.NoError(t, review.Skip())
		nextCard, _ = review.CurrentCard()
		assert.Equal(t, card, nextCard)
	})
}
