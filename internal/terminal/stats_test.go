package terminal_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestCardStats(t *testing.T) {
	t.Parallel()

	t.Run("show card with no stats", func(t *testing.T) {
		m, _ := newTestModel(singleCardDeck).
			Init().
			SendMsg(windowSizeMsg).
			SendKeyType(tea.KeyEnter).
			SendKeyRune(statsKey).
			Get()

		view := m.View()

		assert.Contains(t, view, "No stats")
	})

	t.Run("show errors when find stats fail", func(t *testing.T) {
		m, _ := newTestModel(invalidStats).
			Init().
			SendMsg(windowSizeMsg).
			SendKeyType(tea.KeyEnter).
			SendKeyRune(statsKey).
			Get()

		assert.Contains(t, m.View(), "Error")
	})
}
