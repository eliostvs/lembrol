package terminal_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestCardStats(t *testing.T) {
	t.Parallel()

	t.Run(
		"shows card with no stats", func(t *testing.T) {
			m, _ := newTestModel(singleCardDeck).
				Init().
				SendMsg(windowSizeMsg).
				SendKeyType(tea.KeyEnter).
				SendKeyRune(statsKey).
				Get()

			view := m.View()

			assert.Contains(t, view, "No stats")
			assert.Contains(t, view, "q quit")
		},
	)

	t.Run(
		"shows error when find stats fail", func(t *testing.T) {
			m, _ := newTestModel(invalidStats).
				Init().
				SendMsg(windowSizeMsg).
				SendKeyType(tea.KeyEnter).
				SendKeyRune(statsKey).
				Get()

			assert.Contains(t, m.View(), "Error")
		},
	)

	t.Run(
		"shows card with stats", func(t *testing.T) {
			m, _ := newTestModel(fewDecks).
				Init().
				SendMsg(windowSizeMsg).
				SendKeyType(tea.KeyEnter).
				SendKeyRune(statsKey).
				Print().
				Get()

			view := m.View()

			assert.Contains(t, view, "Stats")
			assert.Contains(t, view, "Question A")
			assert.Contains(t, view, "01/02/2022")
			assert.Contains(t, view, "28/02/2022")
			assert.Contains(t, view, "TOTAL          HARD           NORMAL         EASY           VERY EASY")
			assert.Contains(t, view, "21             5              6              5              5")
			assert.Contains(t, view, "▃▅▃▅▁█▅▅▁▃▁▁▅█▁█▃▃██▃")
			assert.Contains(t, view, "q quit")
		},
	)
}
