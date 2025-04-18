package tui_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	t.Parallel()

	t.Run(
		"shows error page", func(t *testing.T) {
			view := newTestModel(t, invalidDeck).
				Init().
				Get().
				View()

			assert.Contains(t, view, "Error")
			assert.Contains(t, view, "looking for beginning of object key string")
			assert.Contains(t, view, "q quit")
		},
	)

	t.Run("quits the app", func(t *testing.T) {
		view := newTestModel(t, invalidDeck).
			Init().
			SendKeyRune(quitKey).
			Get().
			View()

		assert.NotContains(t, view, "Error")
		assert.Contains(t, view, "Bye")
	})
}
