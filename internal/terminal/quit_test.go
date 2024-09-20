package terminal_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuit(t *testing.T) {
	t.Parallel()

	t.Run("show goodbye screen ", func(t *testing.T) {
		view := newTestModel(t, manyDecks).
			Init().
			SendKeyRune(quitKey).
			Get().
			View()

		assert.Contains(t, view, "Bye")
		assert.Contains(t, view, "Thanks for using Lembrol!")
	},
	)
}
