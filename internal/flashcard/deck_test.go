package flashcard_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eliostvs/remembercli/internal/flashcard"
	"github.com/eliostvs/remembercli/internal/test"
)

var (
	smallDeck   = "Golang Small"
	largeDeck   = "Golang Large"
	oneCardDeck = "Golang One"
)

func TestOpenDeck(t *testing.T) {
	t.Run("returns error when deck file is invalid", func(t *testing.T) {
		location := t.TempDir() + "/foo.toml"
		if err := os.WriteFile(location, []byte("{}"), 0444); err != nil {
			t.Fatal(err)
		}

		repo, err := flashcard.OpenDeck(location, nil)

		assert.Error(t, err)
		assert.Nil(t, repo)
	})
}

func TestDeck_DueCards(t *testing.T) {
	tests := []struct {
		name string
		args time.Time
		deck string
		want int
	}{
		{
			name: "one day after oldest card",
			args: afterOldestCard,
			want: 1,
		},
		{
			name: "two days after oldest card",
			args: afterOldestCard.Add(2 * time.Hour * 24),
			want: 3,
		},
		{
			name: "four days after oldest card",
			args: afterOldestCard.Add(4 * time.Hour * 24),
			want: 5,
		},
		{
			name: "six days after oldest card",
			args: afterOldestCard.Add(6 * time.Hour * 24),
			want: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deck := newDeck(t, largeDeck, withTestClock(tt.args))

			assert.Len(t, deck.DueCards(), tt.want)
		})
	}
}

func TestDeck_Total(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			deck := newDeck(t, tt.name)

			assert.Equal(t, tt.want, deck.Total())
		})
	}
}

func TestDeck_HasDueCards(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			deck := newDeck(t, largeDeck, withTestClock(tt.args))

			assert.Equal(t, tt.want, deck.HasDueCards())
		})
	}
}

func TestDeck_Remove(t *testing.T) {
	t.Run("returns error when parameter is missing", func(t *testing.T) {
		deck := newDeck(t, smallDeck)

		err := deck.Remove(flashcard.Card{})

		assert.ErrorIs(t, err, flashcard.ErrCardNotExist)
	})

	t.Run("removes card from deck", func(t *testing.T) {
		deck := newDeck(t, smallDeck)
		card := deck.List()[0]
		total := deck.Total()

		err := deck.Remove(card)

		assert.NoError(t, err)
		assert.Equal(t, total-1, deck.Total())
		assert.NotContains(t, deck.List(), card)
	})
}

func TestDeck_List(t *testing.T) {
	repo := newRepository(t, "./testdata/sort")
	deck, _ := repo.Open("Sort")

	questions := make([]string, 0, deck.Total())
	for _, card := range deck.List() {
		questions = append(questions, card.Question)
	}

	assert.Equal(t, []string{"A", "B", "C", "D"}, questions)
}

func TestDeck_Add(t *testing.T) {
	now := time.Now()
	repo := newRepository(t, t.TempDir(), withTestClock(now))
	deck, _ := repo.Create("currentDeck")

	card := deck.Add("Question", "Answer")

	assert.Equal(t, 1, deck.Total())
	require.NotNil(t, card)
	assert.Equal(t, card.Question, "Question")
	assert.Equal(t, card.Answer, "Answer")
	assert.Equal(t, card.ReviewedAt, now)
}

func TestDeck_Update(t *testing.T) {
	deck := newDeck(t, "Empty", withDeck("./testdata/empty"))
	card := deck.Add("Question", "Answer")

	card.Question = "Not Question"
	deck.Update(card)

	assert.ElementsMatch(t, deck.List(), []flashcard.Card{card})
}

// Test Options & Factories

type configOption func(*option)

func withDeck(location string) configOption {
	return func(o *option) {
		o.decksLocation = location
	}
}

func withTestClock(t time.Time) configOption {
	return func(o *option) {
		o.clock = test.NewClock(t)
	}
}

type option struct {
	clock         flashcard.Clock
	decksLocation string
}

func newDeck(t *testing.T, deckName string, cfgOpts ...configOption) *flashcard.Deck {
	t.Helper()

	opts := option{
		decksLocation: manyDecksLocation,
		clock:         flashcard.NewClock(),
	}
	for _, cfg := range cfgOpts {
		cfg(&opts)
	}

	repo, err := flashcard.NewRepository(test.TempDirCopy(t, opts.decksLocation), opts.clock)
	if err != nil {
		t.Fatal(err)
	}

	deck, err := repo.Open(deckName)
	if err != nil {
		t.Fatal(err)
	}

	return deck
}
