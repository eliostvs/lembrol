package flashcard_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eliostvs/lembrol/internal/clock"
	testclock "github.com/eliostvs/lembrol/internal/clock/test"
	"github.com/eliostvs/lembrol/internal/flashcard"
)

func TestNewDeck(t *testing.T) {
	t.Parallel()

	type args struct {
		clock clock.Clock
		name  string
		cards []flashcard.Card
	}
	tests := []struct {
		name string
		args
		err error
	}{
		{
			name: "valid deck",
			args: args{
				clock: clock.New(),
				name:  "ValidName",
				cards: []flashcard.Card{
					{ID: "id"},
				},
			},
		},
		{
			name: "missing name",
			args: args{
				clock: clock.New(),
			},
			err: errors.New("missing name"),
		},
		{
			name: "missing clock",
			args: args{
				name: "ValidName",
			},
			err: errors.New("missing clock"),
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				deck, err := flashcard.NewDeck(tt.args.name, tt.args.clock, tt.args.cards)

				if tt.err != nil {
					assert.EqualError(t, err, tt.err.Error())
					assert.Empty(t, deck)
				} else {
					assert.NotEmpty(t, deck.ID)
					assert.Equal(t, tt.args.name, deck.Name)
					assert.Equal(t, tt.args.cards, deck.Cards)
				}
			},
		)
	}
}

func TestDeck_DueCards(t *testing.T) {
	t.Parallel()

	t.Run(
		"empty deck", func(t *testing.T) {
			assert.Equal(t, []flashcard.Card{}, flashcard.Deck{}.DueCards())
		},
	)

	t.Run(
		"non empty deck", func(t *testing.T) {
			tests := []struct {
				name string
				args time.Time
				deck string
				want []string
			}{
				{
					name: "one day after oldest card",
					args: afterOldestCard,
					want: []string{
						"How do you delete a file?",
					},
				},
				{
					name: "two days after oldest card",
					args: afterOldestCard.Add(2 * time.Hour * 24),
					want: []string{
						"How do you check if a map has a certain key?",
						"How do you create a directory?",
						"How do you delete a file?",
					},
				},
				{
					name: "four days after oldest card",
					args: afterOldestCard.Add(4 * time.Hour * 24),
					want: []string{
						"How do you read a whole file?",
						"How do you sleep for x seconds?",
						"How do you check if a map has a certain key?",
						"How do you create a directory?",
						"How do you delete a file?",
					},
				},
				{
					name: "six days after oldest card",
					args: afterOldestCard.Add(6 * time.Hour * 24),
					want: []string{
						"How do you sort an array of ints?",
						"How do you write a string to a file?",
						"How do you read a whole file?",
						"How do you sleep for x seconds?",
						"How do you check if a map has a certain key?",
						"How do you create a directory?",
						"How do you delete a file?",
					},
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name, func(t *testing.T) {
						deck := newTestDeck(t, largeDeck, testclock.New(tt.args))

						assertQuestions(t, tt.want, deck.DueCards())
					},
				)
			}
		},
	)
}

func TestDeck_Total(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args string
		want int
	}{
		{
			name: smallDeck,
			want: 3,
		},
		{
			name: largeDeck,
			want: 7,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				deck := newTestDeck(t, tt.name, clock.New())

				assert.Equal(t, tt.want, deck.Total())
			},
		)
	}
}

func TestDeck_HasDueCards(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args time.Time
		want bool
	}{
		{
			name: "deck has not due cards",
			args: beforeOldestCard,
			want: false,
		},
		{
			name: "deck has due cards",
			args: afterOldestCard,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				deck := newTestDeck(t, largeDeck, testclock.New(tt.args))

				assert.Equal(t, tt.want, deck.HasDueCards())
			},
		)
	}
}

func TestDeck_Remove(t *testing.T) {
	t.Parallel()

	deck := newTestDeck(t, smallDeck, clock.New())
	card := deck.List()[0]
	total := deck.Total()

	newDeck := deck.Remove(card)

	assert.Equal(t, total-1, newDeck.Total())
	assert.NotContains(t, newDeck.List(), card)
}

func TestDeck_List(t *testing.T) {
	t.Parallel()

	deck := newTestDeck(t, "sort.json", clock.New())

	assertQuestions(t, []string{"A", "B", "C", "D"}, deck.List())
}

func TestDeck_Add(t *testing.T) {
	t.Parallel()

	now := time.Now()
	deck := newTestDeck(t, emptyDeck, testclock.New(now))

	newDeck, card := deck.Add("Question", "Answer")

	assert.Equal(t, 0, deck.Total())
	assert.Empty(t, deck.Cards)

	assert.Equal(t, 1, newDeck.Total())
	assert.ElementsMatch(t, newDeck.Cards, []flashcard.Card{card})

	require.NotNil(t, card)
	assert.Equal(t, card.Question, "Question")
	assert.Equal(t, card.Answer, "Answer")
	assert.Equal(t, card.LastReview, now)
}

func TestDeck_Update(t *testing.T) {
	t.Parallel()

	deck := newTestDeck(t, emptyDeck, clock.New())
	deck, card := deck.Add("Question", "Answer")

	card.Question = "Not Question"
	newDeck := deck.Change(card)

	assert.ElementsMatch(t, newDeck.List(), []flashcard.Card{card})
}

/*
 Test Utilities
*/

var (
	emptyDeck  = "empty.json"
	smallDeck  = "small.json"
	largeDeck  = "large.json"
	singleDeck = "single.json"
)

func newTestDeck(t *testing.T, file string, c clock.Clock) flashcard.Deck {
	t.Helper()

	content, err := os.ReadFile(filepath.Join("./testdata", file))
	if err != nil {
		t.Fatal(err)
	}

	var deck flashcard.Deck
	if err := json.Unmarshal(content, &deck); err != nil {
		t.Fatal(err)
	}

	deck, err = flashcard.NewDeck(deck.Name, c, deck.Cards)
	if err != nil {
		t.Fatal(err)
	}

	return deck
}

func assertQuestions(t *testing.T, expected []string, cards []flashcard.Card) {
	t.Helper()

	actual := make([]string, 0, len(cards))
	for _, card := range cards {
		actual = append(actual, card.Question)
	}

	assert.Equal(t, expected, actual)
}
