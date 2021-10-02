package terminal_test

import (
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestLoading(t *testing.T) {
	t.Run("goes to home page when loading is successful", func(t *testing.T) {
		m, _ := newTestModel(manyDecks).
			Init().
			Get()

		assert.Contains(t, m.View(), "Decks")
	})

	t.Run("goes to error page when loading fails", func(t *testing.T) {
		m, _ := newTestModel(invalidDeck).
			Init().
			Get()

		assert.Contains(t, m.View(), "Error")
	})

	t.Run("shows loading page", func(t *testing.T) {
		newTestModel(manyDecks).
			Peek(func(m tea.Model) {
				assert.Contains(t, m.View(), "Remember")
				assert.Contains(t, m.View(), "⣾  Loading...")
			}).
			ForceUpdate(spinner.TickMsg{}).
			Peek(func(m tea.Model) {
				assert.Contains(t, m.View(), "⣽  Loading...")
			}).
			ForceUpdate(spinner.TickMsg{}).
			Peek(func(m tea.Model) {
				assert.Contains(t, m.View(), "⣻  Loading...")
			}).
			ForceUpdate(spinner.TickMsg{}).
			Peek(func(m tea.Model) {
				assert.Contains(t, m.View(), "⢿  Loading...")
			})
	})
}
