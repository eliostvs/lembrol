package tui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestCardStats(t *testing.T) {
	t.Parallel()

	t.Run(
		"shows card with no stats", func(t *testing.T) {
			view := newTestModel(t, singleCardDeck).
				Init().
				SendKeyType(tea.KeyEnter).
				SendKeyType(tea.KeyEnter).
				Get().
				View()

			assert.Contains(t, view, "No stats")
			assert.Contains(t, view, "q quit")
		},
	)

	t.Run(
		"shows card with stats", func(t *testing.T) {
			view := newTestModel(t, fewDecks).
				Init().
				SendKeyType(tea.KeyEnter).
				SendKeyType(tea.KeyEnter).
				Get().
				View()

			assert.Contains(t, view, "Stats")
			assert.Contains(t, view, "Question A")
			assert.Contains(t, view, "01/02/2022")
			assert.Contains(t, view, "28/02/2022")
			assert.Contains(t, view, "TOTAL           VERY EASY       EASY            NORMAL          HARD")
			assert.Contains(t, view, "21              5               6               5")
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
						view := newTestModel(t, fewDecks).
							Init().
							SendKeyType(tea.KeyEnter).
							SendKeyType(tea.KeyEnter).
							SendKeyRune(tt.key).
							Get().
							View()

						assert.NotContains(t, view, "Stats")
					},
				)
			}
		},
	)
}
