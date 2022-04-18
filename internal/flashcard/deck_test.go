package flashcard_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eliostvs/lembrol/internal/flashcard"
	"github.com/eliostvs/lembrol/internal/test"
)

func TestOpenDeck(t *testing.T) {
	t.Run(
		"returns error when deck file is invalid", func(t *testing.T) {
			location := t.TempDir() + "/foo.toml"
			if err := os.WriteFile(location, []byte("{}"), 0444); err != nil {
				t.Fatal(err)
			}

			_, err := flashcard.ReadDeck(location, nil)

			assert.Error(t, err)
		},
	)
}

func TestDeck_DueCards(t *testing.T) {
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
				t.Run(
					tt.name, func(t *testing.T) {
						deck := newDeck(t, largeDeck, withTestClock(tt.args))

						assert.Len(t, deck.DueCards(), tt.want)
					},
				)
			}
		},
	)
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
		t.Run(
			tt.name, func(t *testing.T) {
				deck := newDeck(t, tt.name)

				assert.Equal(t, tt.want, deck.Total())
			},
		)
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
		t.Run(
			tt.name, func(t *testing.T) {
				deck := newDeck(t, largeDeck, withTestClock(tt.args))

				assert.Equal(t, tt.want, deck.HasDueCards())
			},
		)
	}
}

func TestDeck_Remove(t *testing.T) {
	t.Run(
		"returns error when parameter is missing", func(t *testing.T) {
			deck := newDeck(t, smallDeck)

			_, err := deck.Remove(flashcard.Card{})

			assert.ErrorIs(t, err, flashcard.ErrCardNotExist)
		},
	)

	t.Run(
		"removes card from deck", func(t *testing.T) {
			deck := newDeck(t, smallDeck)
			card := deck.List()[0]
			total := deck.Total()

			newDeck, err := deck.Remove(card)

			assert.NoError(t, err)
			assert.Equal(t, total-1, newDeck.Total())
			assert.NotContains(t, newDeck.List(), card)
		},
	)
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
	deck, _ := repo.Create("deck")

	newDeck, card := deck.Add("Question", "Answer")

	assert.Equal(t, 1, newDeck.Total())
	require.NotNil(t, card)
	assert.Equal(t, card.Question, "Question")
	assert.Equal(t, card.Answer, "Answer")
	assert.Equal(t, card.ReviewedAt, now)
}

func TestDeck_Update(t *testing.T) {
	deck := newDeck(t, "Empty", withLocation(test.TempDirCopy(t, "./testdata/empty")))
	newDeck, card := deck.Add("Question", "Answer")

	card.Question = "Not Question"
	deck.Change(card)

	assert.ElementsMatch(t, newDeck.List(), []flashcard.Card{card})
}

func TestNewDeckRepository(t *testing.T) {
	t.Run(
		"returns error deck format is invalid", func(t *testing.T) {
			repo, err := flashcard.NewDeckRepository(invalidDeckLocation, nil)

			assert.Error(t, err)
			assert.Nil(t, repo)
		},
	)

	t.Run(
		"returns repository when the decks location is empty", func(t *testing.T) {
			repo, err := flashcard.NewDeckRepository(t.TempDir(), nil)

			assert.NoError(t, err)
			assert.NotNil(t, repo)
		},
	)

	t.Run(
		"returns repository when the location does not exist", func(t *testing.T) {
			repo, err := flashcard.NewDeckRepository(t.TempDir()+"/foo", nil)

			assert.NoError(t, err)
			assert.NotNil(t, repo)
		},
	)

	t.Run(
		"returns error when create repository fails", func(t *testing.T) {
			location := t.TempDir() + "/foo"
			if err := os.Mkdir(location, 0444); err != nil {
				t.Fatal(err)
			}

			repo, err := flashcard.NewDeckRepository(location+"/bar", nil)

			assert.Error(t, err)
			assert.Nil(t, repo)
		},
	)
}

func TestDeckRepository_List(t *testing.T) {
	t.Run(
		"returns all decks in repository order by name", func(t *testing.T) {
			repo := newRepository(t, manyDecksLocation)

			got := repo.List()

			assert.Len(t, got, 3)
			deckNames := make([]string, 0, len(got))
			for _, deck := range got {
				deckNames = append(deckNames, deck.Name)
			}
			assert.ElementsMatch(t, []string{"Large", "Single", "Small"}, deckNames)
		},
	)

	t.Run(
		"returns empty slice when repository has not decks", func(t *testing.T) {
			repo := newRepository(t, t.TempDir())

			assert.Empty(t, repo.List())
		},
	)
}

func TestDeckRepository_Open(t *testing.T) {
	t.Run(
		"returns err when deck not exist", func(t *testing.T) {
			repo := newRepository(t, manyDecksLocation)

			_, err := repo.Open("Not Found")

			assert.ErrorIs(t, err, flashcard.ErrDeckNotExist)
		},
	)

	t.Run(
		"returns deck", func(t *testing.T) {
			repo, err := flashcard.NewDeckRepository(manyDecksLocation, nil)
			require.NoError(t, err)

			deck, err := repo.Open("Small")

			assert.NoError(t, err)
			assert.Equal(t, 3, deck.Total())
			assert.Len(t, deck.List(), 3)
		},
	)
}

func TestDeckRepository_Save(t *testing.T) {
	t.Run(
		"persist deck changes", func(t *testing.T) {
			location := test.TempDirCopy(t, manyDecksLocation)
			repo := newRepository(t, location)
			originalDeck, _ := repo.Open(smallDeck)

			originalDeck.Name = "Foo"
			err := repo.Save(originalDeck)

			assert.NoError(t, err)
			newDeck, _ := flashcard.ReadDeck(filepath.Join(location, "small.toml"), nil)
			assert.Equal(t, originalDeck.Name, newDeck.Name)
		},
	)

	t.Run(
		"returns error when deck name is invalid", func(t *testing.T) {
			repo := newRepository(t, manyDecksLocation)

			err := repo.Save(flashcard.Deck{})

			assert.EqualError(t, err, "invalid empty file name")
		},
	)

	t.Run(
		"persists new cards", func(t *testing.T) {
			location := test.TempDirCopy(t, emptyDeckLocation)
			repo := newRepository(t, location)
			originalDeck, _ := repo.Open("Empty")
			_, card := originalDeck.Add("question", "answer")

			err := repo.Save(originalDeck)

			assert.NoError(t, err)
			newDeck, _ := flashcard.ReadDeck(filepath.Join(location, "empty.toml"), nil)
			assert.Equal(t, 1, newDeck.Total())
			assert.Equal(t, card.Question, newDeck.List()[0].Question)
			assert.Equal(t, card.Answer, newDeck.List()[0].Answer)
		},
	)
}

func TestDeckRepository_Create(t *testing.T) {
	t.Run(
		"creates deck", func(t *testing.T) {
			location := t.TempDir()
			repo := newRepository(t, location)

			deck, err := repo.Create("Foo Bar")

			assert.NoError(t, err)
			assert.Len(t, repo.List(), 1)
			d, err := flashcard.ReadDeck(filepath.Join(location, "foo-bar.toml"), flashcard.NewClock())
			require.NoError(t, err)
			assert.Equal(t, deck.Name, d.Name)
		},
	)

	t.Run(
		"returns error when persist fails", func(t *testing.T) {
			location, cleanup := test.TempReadOnlyDir(t)
			t.Cleanup(cleanup)
			repo := newRepository(t, location)

			_, err := repo.Create("Foo Bar")
			assert.Error(t, err)
		},
	)
}

func TestDeckRepository_Remove(t *testing.T) {
	t.Run(
		"removes deck from repository", func(t *testing.T) {
			location := test.TempDirCopy(t, manyDecksLocation)
			repo := newRepository(t, location)
			deck, _ := repo.Open(smallDeck)

			err := repo.Remove(deck)

			assert.NoError(t, err)
			_, err = os.Stat(filepath.Join(location, "small.toml"))
			assert.ErrorIs(t, err, os.ErrNotExist)

			_, err = repo.Open(smallDeck)
			assert.ErrorIs(t, err, flashcard.ErrDeckNotExist)
		},
	)

	t.Run(
		"return error when remove resource fails", func(t *testing.T) {
			location, cleanup := test.TempReadOnlyDirCopy(t, manyDecksLocation)
			t.Cleanup(cleanup)
			repo := newRepository(t, location)
			deck, _ := repo.Open(smallDeck)

			err := repo.Remove(deck)

			assert.Error(t, err)
		},
	)

	t.Run(
		"returns error when deck is not found", func(t *testing.T) {
			repo := newRepository(t, manyDecksLocation)

			err := repo.Remove(flashcard.Deck{})

			assert.ErrorIs(t, err, flashcard.ErrDeckNotExist)
		},
	)
}

func TestDeckRepository_Total(t *testing.T) {
	tests := []struct {
		name string
		args string
		want int
	}{
		{
			name: "three decks",
			args: manyDecksLocation,
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				repo := newRepository(t, tt.args)

				assert.Equal(t, repo.Total(), tt.want)
			},
		)
	}
}

/*
 Test Utilities
*/

var (
	smallDeck           = "Small"
	largeDeck           = "Large"
	singleCardDeck      = "Single"
	emptyDeck           = "Empty"
	manyDecksLocation   = "./testdata/many"
	emptyDeckLocation   = "./testdata/empty"
	invalidDeckLocation = "./testdata/invalid-deck"
)

type configOption func(*option)

func withLocation(location string) configOption {
	return func(o *option) {
		o.location = location
	}
}

func withTestClock(t time.Time) configOption {
	return func(o *option) {
		o.clock = test.NewClock(t)
	}
}

type option struct {
	clock    flashcard.Clock
	location string
}

func newDeck(t *testing.T, deckName string, cfgOpts ...configOption) flashcard.Deck {
	t.Helper()

	opts := option{
		location: test.TempDirCopy(t, manyDecksLocation),
		clock:    flashcard.NewClock(),
	}
	for _, cfg := range cfgOpts {
		cfg(&opts)
	}

	repo, err := flashcard.NewDeckRepository(opts.location, opts.clock)
	if err != nil {
		t.Fatal(err)
	}

	deck, err := repo.Open(deckName)
	if err != nil {
		t.Fatal(err)
	}

	return deck
}

func newRepository(t *testing.T, deckLocation string, cfgOpts ...configOption) *flashcard.DeckRepository {
	t.Helper()

	opts := option{
		clock: flashcard.NewClock(),
	}
	for _, cfg := range cfgOpts {
		cfg(&opts)
	}

	repo, err := flashcard.NewDeckRepository(deckLocation, opts.clock)
	if err != nil {
		t.Fatal(err)
	}

	return repo
}
