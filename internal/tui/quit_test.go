package tui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/eliostvs/lembrol/internal/tui"
	"github.com/stretchr/testify/assert"
)

func TestQuit(t *testing.T) {
	t.Parallel()

	t.Run("shows goodbye screen", func(t *testing.T) {
		view := newTestModel(t, manyDecks).
			Init().
			SendKeyRune(quitKey).
			Get().
			View()

		assert.Contains(t, view, "Bye")
		assert.Contains(t, view, "Thanks for using Lembrol!")
	},
	)

	t.Run(
		"changes the height when the window resize", func(t *testing.T) {
			var before string

			after := newTestModel(t, emptyDeck, tui.WithWindowSize(testWidth, testHeight*2)).
				Init().
				SendKeyRune(quitKey).
				Peek(
					func(m tea.Model) {
						before = m.View()
					},
				).
				SendMsg(tea.WindowSizeMsg{Width: testWidth, Height: testHeight / 2}).
				Get().
				View()

			assert.NotEqual(t, before, after)
		},
	)
}
