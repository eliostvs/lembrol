package tui_test

import (
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/eliostvs/lembrol/internal/tui"
)

func TestLoading(t *testing.T) {
	t.Parallel()

	t.Run(
		"goes to home page when loading is successful", func(t *testing.T) {
			view := newTestModel(t, manyDecks).
				Init().
				Get().
				View()

			assert.Contains(t, view, "Decks")
		},
	)

	t.Run(
		"goes to error page when loading fails", func(t *testing.T) {
			view := newTestModel(t, invalidDeck).
				Init().
				Get().
				View()

			assert.Contains(t, view, "Error")
		},
	)

	t.Run(
		"shows loading page", func(t *testing.T) {
			view := newTestModel(t, manyDecks).Get().View()

			assert.Contains(t, view, "Loading...")
			assert.Contains(t, view, "ctrl+c quit")
		},
	)

	t.Run(
		"quits apps", func(t *testing.T) {
			view := newTestModel(t, manyDecks).
				SendKeyRune("ctrl+c").
				Get().
				View()

			assert.Contains(t, view, "Thanks for using Lembrol!")
		},
	)

	t.Run(
		"changes the layout when the window resize", func(t *testing.T) {
			var before string

			after := newTestModel(t, emptyDeck, tui.WithWindowSize(testWidth, testHeight*2)).
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

	t.Run(
		"changes loading animation", func(t *testing.T) {
			var before string

			after := newTestModel(t, manyDecks).
				Peek(
					func(m tea.Model) {
						before = m.View()
					},
				).
				ForceUpdate(spinner.TickMsg{}).
				Get().
				View()

			assert.NotEqual(t, before, after)
		},
	)
}
