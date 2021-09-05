package terminal_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	t.Run("shows error page", func(t *testing.T) {
		m, _ := newTestModel(invalidDeck).
			init().
			Get()

		view := m.View()

		assert.Contains(t, view, "Error")
		assert.Contains(t, view, "unmarshall deck")
	})
}
