package flashcard_test

import (
	"os"
	"testing"
	"time"

	"github.com/eliostvs/lembrol/internal/clock"
	"github.com/eliostvs/lembrol/internal/flashcard"
	"github.com/eliostvs/lembrol/internal/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	manyDecksPath = "./testdata/many"
	fewDecksPath  = "./testdata/few"
	emptyDeckPath = "./testdata/empty"
)

func TestNewDeckRepository(t *testing.T) {
	t.Run("returns error when a deck file is in invalid format", func(t *testing.T) {
		repo, err := flashcard.NewRepository("./testdata/invalid", nil)

		assert.Error(t, err)
		assert.Nil(t, repo)
	})

	t.Run("returns empty repository when the decks directory is empty", func(t *testing.T) {
		repo, err := flashcard.NewRepository(t.TempDir(), nil)

		assert.NoError(t, err)
		assert.Equal(t, 0, repo.Total())
		assert.NotNil(t, repo)
	})

	t.Run("returns empty repository when the directory does not exist", func(t *testing.T) {
		repo, err := flashcard.NewRepository(t.TempDir()+"/foo", nil)

		assert.NoError(t, err)
		assert.NotNil(t, repo)
	})

	t.Run("returns error when create repository fails", func(t *testing.T) {
		tempDir := t.TempDir() + "/foo"
		if err := os.Mkdir(tempDir, 0o444); err != nil {
			t.Fatal(err)
		}

		repo, err := flashcard.NewRepository(tempDir+"/bar", nil)

		assert.Error(t, err)
		assert.Nil(t, repo)
	})
}

func TestDeckRepository_Total(t *testing.T) {
	tests := []struct {
		name string
		args string
		want int
	}{
		{
			name: "no decks",
			args: emptyDeckPath,
			want: 0,
		},
		{
			name: "many decks",
			args: manyDecksPath,
			want: 6,
		},
		{
			name: "few decks",
			args: fewDecksPath,
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newTestRepository(t, tt.args, nil)

			assert.Equal(t, tt.want, repo.Total())
		})
	}
}

func TestDeckRepository_List(t *testing.T) {
	tests := []struct {
		name string
		args string
		want []string
	}{
		{
			name: "no decks",
			args: emptyDeckPath,
		},
		{
			name: "many decks",
			args: manyDecksPath,
			want: []string{"Golang A", "Golang B", "Golang C", "Golang D", "Golang E", "Golang F"},
		},
		{
			name: "few decks",
			args: fewDecksPath,
			want: []string{"Golang A", "Golang B"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newTestRepository(t, tt.args, nil)

			assert.ElementsMatch(t, tt.want, deckNames(repo.List()))
		})
	}
}

func TestDeckRepository_Create(t *testing.T) {
	t.Run("creates deck", func(t *testing.T) {
		deckName := test.RandomName()
		repo := newTestRepository(t, t.TempDir(), clock.New())
		card := flashcard.NewCard(test.RandomName(), test.RandomName(), time.Now())

		deck, err := repo.Create(deckName, []flashcard.Card{card})

		assert.NoError(t, err)
		assert.Equal(t, deckName, deck.Name)
	})

	t.Run("returns error when persist fails", func(t *testing.T) {
		location, cleanup := test.TempReadOnlyCopyDir(t, t.TempDir())
		t.Cleanup(cleanup)
		repo := newTestRepository(t, location, clock.New())
		card := flashcard.NewCard(test.RandomName(), test.RandomName(), time.Now())

		deck, err := repo.Create(test.RandomName(), []flashcard.Card{card})
		assert.Empty(t, deck)
		assert.Error(t, err)
	})

	t.Run("returns error when deck is invalid", func(t *testing.T) {
		repo := newTestRepository(t, t.TempDir(), clock.New())
		card := flashcard.NewCard(test.RandomName(), test.RandomName(), time.Now())

		deck, err := repo.Create("", []flashcard.Card{card})
		assert.Empty(t, deck)
		assert.Error(t, err)
	})

	t.Run("returns error when deck is duplicate", func(t *testing.T) {
		repo := newTestRepository(t, t.TempDir(), clock.New())
		deckName := test.RandomName()
		card := flashcard.NewCard(test.RandomName(), test.RandomName(), time.Now())

		_, err := repo.Create(deckName, []flashcard.Card{card})
		require.NoError(t, err)

		deck, err := repo.Create(deckName, []flashcard.Card{card})

		assert.Empty(t, deck)
		assert.Error(t, err)
	})
}

func TestDeckRepository_Delete(t *testing.T) {
	t.Run("deletes deck", func(t *testing.T) {
		tempDir := test.TempCopyDir(t, manyDecksPath)
		repo := newTestRepository(t, tempDir, nil)
		deck := repo.List()[0]

		err := repo.Delete(deck)

		assert.NoError(t, err)

		_, err = os.Stat(deck.ID)
		assert.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("return error when delete file fails", func(t *testing.T) {
		location, cleanup := test.TempReadOnlyCopyDir(t, manyDecksPath)
		t.Cleanup(cleanup)

		repo := newTestRepository(t, location, nil)
		deck := repo.List()[0]

		err := repo.Delete(deck)

		assert.Error(t, err)
	})

	t.Run("returns error when deck is not found", func(t *testing.T) {
		repo := newTestRepository(t, manyDecksPath, nil)

		err := repo.Delete(flashcard.Deck{})

		assert.ErrorIs(t, err, flashcard.ErrDeckNotFound)
	})
}

func TestRepository_Find(t *testing.T) {
	repo := newTestRepository(t, manyDecksPath, nil)
	deck := repo.List()[0]

	type want struct {
		err  error
		deck flashcard.Deck
	}
	tests := []struct {
		name string
		args string
		want want
	}{
		{
			name: "finds the deck",
			args: deck.Name,
			want: want{
				deck: deck,
			},
		},
		{
			name: "don't find the deck",
			args: test.RandomName(),
			want: want{
				deck: flashcard.Deck{},
				err:  flashcard.ErrDeckNotFound,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deck, err := repo.Find(tt.args)

			assert.Equal(t, tt.want.err, err)
			assert.Equal(t, tt.want.deck, deck)
		})
	}
}

/*
func TestDeckRepository_Save(t *testing.T) {
	t.Run("persist deck changes", func(t *testing.T) {
		location := test.TempDirCopy(t, manyDecksLocation)
		repo := newTestRepository(t, location)
		originalDeck, _ := repo.Open(smallDeck)

		originalDeck.Name = "Foo"
		err := repo.Save(originalDeck)

		assert.NoError(t, err)
		newDeck, _ := flashcard.OpenDeck(filepath.Join(location, "small.toml"), nil)
		assert.Equal(t, originalDeck.Name, newDeck.Name)
	})

	t.Run("returns error when parameter is missing", func(t *testing.T) {
		repo := newTestRepository(t, t.TempDir())

		err := repo.Save(nil)

		assert.ErrorIs(t, err, flashcard.ErrDeckNotExist)
	})

	t.Run("returns error when deck is not found", func(t *testing.T) {
		repo := newTestRepository(t, manyDecksLocation)

		err := repo.Save(&flashcard.Deck{})

		assert.ErrorIs(t, err, flashcard.ErrDeckNotExist)
	})

	t.Run("persists new cards", func(t *testing.T) {
		location := test.TempDirCopy(t, emptyDeckLocation)
		repo := newTestRepository(t, location)
		originalDeck, _ := repo.Open("Empty")
		originalDeck.Add("question", "answer")

		err := repo.Save(originalDeck)

		assert.NoError(t, err)
		newDeck, _ := flashcard.OpenDeck(filepath.Join(location, "empty.toml"), nil)
		assert.Equal(t, 1, newDeck.Total())
		assert.Equal(t, "question", newDeck.List()[0].Question)
		assert.Equal(t, "answer", newDeck.List()[0].Answer)
	})
}

func TestDeckRepository_Create(t *testing.T) {
	t.Run("creates deck", func(t *testing.T) {
		location := t.TempDir()
		repo := newTestRepository(t, location)

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
		repo := newTestRepository(t, location)

		deck, err := repo.Create("Foo Bar")
		assert.Nil(t, deck)
		assert.Error(t, err)
	})
}

func TestDeckRepository_Remove(t *testing.T) {
	t.Run("removes deck from repository", func(t *testing.T) {
		location := test.TempDirCopy(t, manyDecksLocation)
		repo := newTestRepository(t, location)
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

		repo := newTestRepository(t, location)
		deck, _ := repo.Open(smallDeck)

		err := repo.Remove(deck)

		assert.Error(t, err)
	})

	t.Run("returns error when parameter is missing", func(t *testing.T) {
		repo := newTestRepository(t, t.TempDir())

		err := repo.Remove(nil)

		assert.ErrorIs(t, err, flashcard.ErrDeckNotExist)
	})

	t.Run("returns error when deck is not found", func(t *testing.T) {
		repo := newTestRepository(t, manyDecksLocation)

		err := repo.Remove(&flashcard.Deck{})

		assert.ErrorIs(t, err, flashcard.ErrDeckNotExist)
	})
}


/*
 Test Utilities
*/

func deckNames(decks []flashcard.Deck) []string {
	names := make([]string, 0, len(decks))

	for _, deck := range decks {
		names = append(names, deck.Name)
	}

	return names
}

func newTestRepository(t *testing.T, path string, clock clock.Clock) *flashcard.Repository {
	t.Helper()

	repo, err := flashcard.NewRepository(path, clock)
	if err != nil {
		t.Fatal(err)
	}

	return repo
}
