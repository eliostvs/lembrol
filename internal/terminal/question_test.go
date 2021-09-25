package terminal_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/eliostvs/remembercli/internal/flashcard"
	"github.com/eliostvs/remembercli/internal/test"
)

func TestQuestion(t *testing.T) {
	t.Run("shows question page", func(t *testing.T) {
		m, _ := newTestModel(singleCardDeck).
			init().
			SendKeyRune(studyKey).
			Get()

		view := m.View()

		assert.Contains(t, view, "Question A")
		assert.Contains(t, view, "Golang One")
		assert.Contains(t, view, "1 of 1")
		assert.Contains(t, view, "enter: answer")
		assert.Contains(t, view, "c: close")
		assert.NotContains(t, view, "s: skip")
		assert.Contains(t, view, "q: quit")
	})

	t.Run("shows skip action", func(t *testing.T) {
		newTestModel(fewDecks).
			init().
			SendKeyRune(studyKey).
			Peek(func(m tea.Model) {
				assert.Contains(t, m.View(), "s: skip")
			})
	})

	t.Run("wraps long answers", func(t *testing.T) {
		width := 70

		m, _ := newTestModel(longNamesDeck).
			init().
			SendKeyRune(studyKey).
			SendMsg(tea.WindowSizeMsg{Width: width, Height: 20}).
			Get()

		view := m.View()

		assert.Contains(t, view, "Very Long Question & Answer")
		assertContainsMarkdown(t, view, width, longestCard.Question)
	})

	t.Run("goes to deck page when the review is canceled", func(t *testing.T) {
		m, _ := newTestModel(manyDecks).
			init().
			SendKeyRune(studyKey).
			SendKeyType(tea.KeyEsc).
			Get()

		assert.Contains(t, m.View(), "Deck")
	})

	t.Run("goes to answer page to answer the question", func(t *testing.T) {
		m, _ := newTestModel(singleCardDeck).
			init().
			SendKeyRune(studyKey).
			SendKeyType(tea.KeyEnter).
			Get()

		assert.Contains(t, m.View(), "Answer")
	})

	t.Run("quits the app from question page", func(t *testing.T) {
		m, _ := newTestModel(test.TempDirCopy(t, singleCardDeck)).
			init().
			SendKeyRune(studyKey).
			SendKeyRune(quitKey).
			Get()

		assert.Contains(t, m.View(), "Thanks for using Remember CLI!")
	})

	t.Run("skips the current card", func(t *testing.T) {
		questionA := "Question A"
		questionB := "Question B"

		choose := func(s string) string {
			if strings.Contains(s, questionA) {
				return questionB
			}
			return questionA
		}
		var nextQuestion string

		newTestModel(test.TempDirCopy(t, fewDecks)).
			init().
			SendKeyType(tea.KeyDown).
			SendKeyRune(studyKey).
			Peek(func(m tea.Model) {
				assert.Contains(t, m.View(), "1 of 2")
				nextQuestion = choose(m.View())
			}).
			SendKeyRune(skipKey).
			Peek(func(m tea.Model) {
				assert.Contains(t, m.View(), "1 of 2")
				assert.Contains(t, m.View(), nextQuestion)
				nextQuestion = choose(m.View())
			}).
			SendKeyRune(skipKey).
			Peek(func(m tea.Model) {
				assert.Contains(t, m.View(), "1 of 2")
				assert.Contains(t, m.View(), nextQuestion)
			})
	})

	t.Run("do not shows skip in the last question", func(t *testing.T) {
		newTestModel(test.TempDirCopy(t, fewDecks)).
			init().
			SendKeyType(tea.KeyDown).
			SendKeyRune(studyKey).
			SendKeyType(tea.KeyEnter).
			SendKeyRune(flashcard.ReviewScoreNormal.String()).
			Peek(func(m tea.Model) {
				assert.Contains(t, m.View(), "2 of 2")
				assert.NotContains(t, m.View(), "skip")
			})
	})
}
