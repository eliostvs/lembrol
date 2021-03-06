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
				SendKeyType(tea.KeyEnter).
				SendKeyRune(statsKey).
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

	t.Run(
		"goes back to card page", func(t *testing.T) {
			tests := []struct {
				name string
				key  string
			}{
				{
					name: "when esc key is pressed",
					key:  tea.KeyEsc.String(),
				},
				{
					name: "when quit key is pressed",
					key:  quitKey,
				},
			}
			for _, tt := range tests {
				t.Run(
					tt.name, func(t *testing.T) {
						m, _ := newTestModel(fewDecks).
							Init().
							SendKeyType(tea.KeyEnter).
							SendKeyRune(statsKey).
							SendKeyRune(tt.key).
							Get()

						assert.NotContains(t, m.View(), "▃▅▃▅▁█▅▅▁▃▁▁▅█▁█▃▃██▃")
					},
				)
			}
		},
	)
}
