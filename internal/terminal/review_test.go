package terminal_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/eliostvs/remembercli/internal/flashcard"
	"github.com/eliostvs/remembercli/internal/test"
)

func TestReview(t *testing.T) {
	t.Run("shows review page", func(t *testing.T) {
		m, _ := newTestModel(test.TempDirCopy(t, singleCardDeck)).
			init().
			SendKeyRune(studyKey).
			SendKeyType(tea.KeyEnter).
			SendKeyRune(flashcard.ReviewScoreNormal.String()).
			Get()

		view := m.View()

		assert.Contains(t, view, "Congratulations!")
		assert.Contains(t, view, "1 card reviewed")
		assert.Contains(t, view, "c: close")
		assert.Contains(t, view, "q: quit")
	})

	t.Run("goes to home page when review is closed", func(t *testing.T) {
		m, _ := newTestModel(test.TempDirCopy(t, singleCardDeck)).
			init().
			SendKeyRune(studyKey).
			SendKeyType(tea.KeyEnter).
			SendKeyRune(flashcard.ReviewScoreNormal.String()).
			SendKeyType(tea.KeyEsc).
			Get()

		assert.Contains(t, m.View(), "Decks")
	})

	t.Run("quits the app from the review page", func(t *testing.T) {
		m, _ := newTestModel(test.TempDirCopy(t, singleCardDeck)).
			init().
			SendKeyRune(studyKey).
			SendKeyType(tea.KeyEnter).
			SendKeyRune(flashcard.ReviewScoreNormal.String()).
			SendKeyRune(quitKey).
			Get()

		assert.Contains(t, m.View(), "Thanks for using Remember CLI")
	})
}
