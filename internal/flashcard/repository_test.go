package flashcard_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eliostvs/lembrol/internal/flashcard"
)

func TestNewRepository(t *testing.T) {
	t.Run("returns error when inittialization of the wrapped repositories fails", func(t *testing.T) {
		location := t.TempDir() + "/foo"
		if err := os.Mkdir(location, 0444); err != nil {
			t.Fatal(err)
		}

		repo, err := flashcard.NewRepository(location+"/foo", flashcard.NewClock())

		assert.Nil(t, repo)
		assert.Error(t, err)
	})

	t.Run("return repository when initializing succeed", func(t *testing.T) {
		repo, err := flashcard.NewRepository(t.TempDir(), flashcard.NewClock())

		assert.NotNil(t, repo)
		assert.NoError(t, err)
		assert.IsType(t, repo.Deck, &flashcard.DeckRepository{})
		assert.IsType(t, repo.Stats, &flashcard.StatsRepository{})
	})
}
