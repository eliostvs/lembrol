package flashcard_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eliostvs/remembercli/internal/flashcard"
	"github.com/eliostvs/remembercli/internal/test"
)

func TestNewDeckRepository(t *testing.T) {
	t.Run("returns error deck format is invalid", func(t *testing.T) {
		repo, err := flashcard.NewDeckRepository(invalidDeckLocation, nil)

		assert.Error(t, err)
		assert.Nil(t, repo)
	})

	t.Run("returns repository when the decks location is empty", func(t *testing.T) {
		repo, err := flashcard.NewDeckRepository(t.TempDir(), nil)

		assert.NoError(t, err)
		assert.NotNil(t, repo)
	})

	t.Run("returns repository when the location does not exist", func(t *testing.T) {
		repo, err := flashcard.NewDeckRepository(t.TempDir()+"/foo", nil)

		assert.NoError(t, err)
		assert.NotNil(t, repo)
	})

	t.Run("returns error when create repository fails", func(t *testing.T) {
		location := t.TempDir() + "/foo"
		if err := os.Mkdir(location, 0444); err != nil {
			t.Fatal(err)
		}

		repo, err := flashcard.NewDeckRepository(location+"/bar", nil)

		assert.Error(t, err)
		assert.Nil(t, repo)
	})
}

func TestDeckRepository_List(t *testing.T) {
	t.Run("returns all decks in repository order by name", func(t *testing.T) {
		repo := newRepository(t, manyDecksLocation)

		got := repo.List()

		assert.Len(t, got, 3)
		deckNames := make([]string, 0, len(got))
		for _, deck := range got {
			deckNames = append(deckNames, deck.Name)
		}
		assert.Equal(t, []string{"Golang Large", "Golang One", "Golang Small"}, deckNames)
	})

	t.Run("returns empty slice when repository has not decks", func(t *testing.T) {
		repo := newRepository(t, t.TempDir())

		assert.Empty(t, repo.List())
	})
}

func TestDeckRepository_Open(t *testing.T) {
	t.Run("returns err when deck not exist", func(t *testing.T) {
		repo := newRepository(t, manyDecksLocation)

		_, err := repo.Open("Not Found")

		assert.ErrorIs(t, err, flashcard.ErrDeckNotExist)
	})

	t.Run("returns deck", func(t *testing.T) {
		repo, err := flashcard.NewDeckRepository(manyDecksLocation, nil)
		require.NoError(t, err)

		deck, err := repo.Open("Golang Small")

		assert.NoError(t, err)
		assert.Equal(t, 3, deck.Total())
		assert.Len(t, deck.List(), 3)
	})
}

func TestDeckRepository_Save(t *testing.T) {
	t.Run("persist deck changes", func(t *testing.T) {
		location := test.TempDirCopy(t, manyDecksLocation)
		repo := newRepository(t, location)
		originalDeck, _ := repo.Open(smallDeck)

		originalDeck.Name = "Foo"
		err := repo.Save(originalDeck)

		assert.NoError(t, err)
		newDeck, _ := flashcard.OpenDeck(filepath.Join(location, "small.toml"), nil)
		assert.Equal(t, originalDeck.Name, newDeck.Name)
	})

	t.Run("returns error when deck is not found", func(t *testing.T) {
		repo := newRepository(t, manyDecksLocation)

		err := repo.Save(flashcard.Deck{})

		assert.ErrorIs(t, err, flashcard.ErrDeckNotExist)
	})

	t.Run("persists new cards", func(t *testing.T) {
		location := test.TempDirCopy(t, emptyDeckLocation)
		repo := newRepository(t, location)
		originalDeck, _ := repo.Open("Empty")
		_, card := originalDeck.Add("question", "answer")

		err := repo.Save(originalDeck)

		assert.NoError(t, err)
		newDeck, _ := flashcard.OpenDeck(filepath.Join(location, "empty.toml"), nil)
		assert.Equal(t, 1, newDeck.Total())
		assert.Equal(t, card.Question, newDeck.List()[0].Question)
		assert.Equal(t, card.Answer, newDeck.List()[0].Answer)
	})
}

func TestDeckRepository_Create(t *testing.T) {
	t.Run("creates deck", func(t *testing.T) {
		location := t.TempDir()
		repo := newRepository(t, location)

		deck, err := repo.Create("Foo Bar")

		assert.NoError(t, err)
		assert.Len(t, repo.List(), 1)
		d, err := flashcard.OpenDeck(filepath.Join(location, "foo-bar.toml"), flashcard.NewClock())
		require.NoError(t, err)
		assert.Equal(t, deck.Name, d.Name)
	})

	t.Run("returns error when persist fails", func(t *testing.T) {
		location, cleanup := test.TempReadOnlyDirCopy(t, t.TempDir())
		t.Cleanup(cleanup)
		repo := newRepository(t, location)

		_, err := repo.Create("Foo Bar")
		assert.Error(t, err)
	})
}

func TestDeckRepository_Remove(t *testing.T) {
	t.Run("removes deck from repository", func(t *testing.T) {
		location := test.TempDirCopy(t, manyDecksLocation)
		repo := newRepository(t, location)
		deck, _ := repo.Open(smallDeck)

		err := repo.Remove(deck)

		assert.NoError(t, err)
		_, err = os.Stat(filepath.Join(location, "small.toml"))
		assert.ErrorIs(t, err, os.ErrNotExist)

		_, err = repo.Open(smallDeck)
		assert.ErrorIs(t, err, flashcard.ErrDeckNotExist)
	})

	t.Run("return error when remove resource fails", func(t *testing.T) {
		location, cleanup := test.TempReadOnlyDirCopy(t, manyDecksLocation)
		t.Cleanup(cleanup)
		repo := newRepository(t, location)
		deck, _ := repo.Open(smallDeck)

		err := repo.Remove(deck)

		assert.Error(t, err)
	})

	t.Run("returns error when deck is not found", func(t *testing.T) {
		repo := newRepository(t, manyDecksLocation)

		err := repo.Remove(flashcard.Deck{})

		assert.ErrorIs(t, err, flashcard.ErrDeckNotExist)
	})
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
		t.Run(tt.name, func(t *testing.T) {
			repo := newRepository(t, tt.args)

			assert.Equal(t, repo.Total(), tt.want)
		})
	}
}

func TestRepository_SaveStats(t *testing.T) {
	t.Run("returns error when save stats fail", func(t *testing.T) {
		location, cleanup := test.TempReadOnlyDirCopy(t, manyDecksLocation)
		t.Cleanup(cleanup)
		repo := newRepository(t, location)
		deck, _ := repo.Open(smallDeck)

		err := repo.SaveStats(deck, &flashcard.Stats{})

		assert.Error(t, err)
	})

	t.Run("save stats to disk", func(t *testing.T) {
		location := test.TempDirCopy(t, manyDecksLocation)
		repo := newRepository(t, location)
		deck, _ := repo.Open(smallDeck)

		stats := flashcard.Stats{
			Algorithm:      "Algorithm",
			Card:           "Card",
			Timestamp:      "Timestamp",
			Score:          "1",
			LastReview:     "LastReview",
			Repetitions:    1,
			Interval:       "Interval",
			EasinessFactor: "EasinessFactor",
		}
		err := repo.SaveStats(deck, &stats)
		assert.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(location, "golang-small-stats.jsonl"))
		require.NoError(t, err)
		want := `{"algorithm":"Algorithm","card":"Card","timestamp":"Timestamp","score":"1","last_review":"LastReview","repetitions":1,"interval":"Interval","easiness_factor":"EasinessFactor"}`
		assert.JSONEq(t, want, string(content))
	})
}

/*
 Test Utilities
*/

var (
	manyDecksLocation   = "./testdata/many"
	emptyDeckLocation   = "./testdata/empty"
	invalidDeckLocation = "./testdata/invalid"
)

func newRepository(t *testing.T, deckLocation string, cfgOpts ...configOption) *flashcard.Repository {
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
