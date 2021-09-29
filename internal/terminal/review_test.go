package terminal_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eliostvs/remembercli/internal/flashcard"
	"github.com/eliostvs/remembercli/internal/test"
)

func TestQuestion(t *testing.T) {
	t.Run("show full options help", func(t *testing.T) {
		m, _ := newTestModel(fewDecks).
			init().
			SendMsg(windowSizeMsg).
			SendKeyRune(studyKey).
			SendKeyRune("?").
			Get()

		view := m.View()

		assert.NotContains(t, view, "enter answer • q quit • ? more")
		assert.Contains(t, view, "s skip    enter answer    q quit")
		assert.Contains(t, view, "? close help")
	})

	t.Run("shows question page", func(t *testing.T) {
		m, _ := newTestModel(singleCardDeck).
			init().
			SendKeyRune(studyKey).
			Get()

		view := m.View()

		assert.Contains(t, view, "Question A")
		assert.Contains(t, view, "Golang One")
		assert.Contains(t, view, "1 of 1")
		assert.Contains(t, view, "enter answer • q quit • ? more")
	})

	t.Run("shows question page with skip action", func(t *testing.T) {
		m, _ := newTestModel(fewDecks).
			init().
			SendKeyRune(studyKey).
			Get()

		view := m.View()

		assert.Contains(t, view, "s skip • enter answer • q quit • ? more")
	})

	t.Run("goes to deck page when the review is canceled", func(t *testing.T) {
		m, _ := newTestModel(manyDecks).
			init().
			SendKeyRune(studyKey).
			SendKeyType(tea.KeyEsc).
			Get()

		assert.Contains(t, m.View(), "6 items")
	})

	t.Run("goes to answer page to answer the question", func(t *testing.T) {
		m, _ := newTestModel(singleCardDeck).
			init().
			SendKeyRune(studyKey).
			SendKeyType(tea.KeyEnter).
			Get()

		assert.Contains(t, m.View(), "Answer")
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
		m, _ := newTestModel(test.TempDirCopy(t, fewDecks)).
			init().
			SendMsg(windowSizeMsg).
			SendKeyType(tea.KeyDown).
			SendKeyRune(studyKey).
			SendKeyType(tea.KeyEnter).
			SendKeyRune(flashcard.ReviewScoreNormal.String()).
			Get()

		view := m.View()

		assert.Contains(t, view, "2 of 2")
		assert.NotContains(t, view, "skip")
	})
}

func TestAnswer(t *testing.T) {
	t.Run("shows answer page", func(t *testing.T) {
		m, _ := newTestModel(singleCardDeck).
			init().
			SendKeyRune(studyKey).
			SendKeyType(tea.KeyEnter).
			Get()

		view := m.View()

		assert.Contains(t, view, "Answer")
		assert.Contains(t, view, "Golang One")
		assert.Contains(t, view, "1 of 1")
		assert.Contains(t, view, latestCard.Answer)
		assert.Contains(t, view, "1 hard • 2 normal • 3 easy • 4 very easy • q quit • ? more")
	})

	t.Run("show full options help", func(t *testing.T) {
		m, _ := newTestModel(singleCardDeck).
			init().
			SendMsg(windowSizeMsg).
			SendKeyRune(studyKey).
			SendKeyType(tea.KeyEnter).
			SendKeyRune("?").
			Get()

		view := m.View()

		assert.NotContains(t, view, "1 hard • 2 normal • 3 easy • 4 very easy • q quit • ? more")
		assert.Contains(t, view, "0 again    1 hard         q quit")
		assert.Contains(t, view, "2 normal       ? close help")
		assert.Contains(t, view, "3 easy")
		assert.Contains(t, view, "4 very easy")
	})

	t.Run("goes to deck page when the review is canceled", func(t *testing.T) {
		m, _ := newTestModel(manyDecks).
			init().
			SendKeyRune(studyKey).
			SendKeyType(tea.KeyEnter).
			SendKeyType(tea.KeyEsc).
			Get()

		assert.Contains(t, m.View(), "6 items")
	})

	t.Run("goes to next review card when the card rating is successful", func(t *testing.T) {
		tests := []struct {
			name string
			args flashcard.ReviewScore
			want string
		}{
			{
				name: "score again",
				args: flashcard.ReviewScoreAgain,
				want: "1 of 6",
			},
			{
				name: "score hard",
				args: flashcard.ReviewScoreHard,
				want: "2 of 6",
			},
			{
				name: "score normal",
				args: flashcard.ReviewScoreNormal,
				want: "2 of 6",
			},
			{
				name: "score easy",
				args: flashcard.ReviewScoreEasy,
				want: "2 of 6",
			},
			{
				name: "score super easy",
				args: flashcard.ReviewScoreSuperEasy,
				want: "2 of 6",
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				m, _ := newTestModel(test.TempDirCopy(t, fewDecks)).
					init().
					SendKeyRune(studyKey).
					SendKeyType(tea.KeyEnter).
					SendKeyRune(tt.args.String()).
					Get()

				assert.Contains(t, m.View(), tt.want)
			})
		}
	})

	t.Run("goes to error page when the rating fails", func(t *testing.T) {
		location := test.TempDirCopy(t, singleCardDeck)
		err := os.Chmod(filepath.Join(location, "single.toml"), 0444)
		require.NoError(t, err)

		m, _ := newTestModel(location).
			init().
			SendKeyRune(studyKey).
			SendKeyType(tea.KeyEnter).
			SendKeyRune(flashcard.ReviewScoreNormal.String()).
			Get()

		assert.Contains(t, m.View(), "Error")
	})

	t.Run("goes to review page when the review ends", func(t *testing.T) {
		m, _ := newTestModel(test.TempDirCopy(t, singleCardDeck)).
			init().
			SendKeyRune(studyKey).
			SendKeyType(tea.KeyEnter).
			SendKeyRune(flashcard.ReviewScoreNormal.String()).
			Get()

		assert.Contains(t, m.View(), "Congratulations")
	})
}

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
		assert.Contains(t, view, "q quit • ? more")
	})

	t.Run("show full options help", func(t *testing.T) {
		m, _ := newTestModel(test.TempDirCopy(t, singleCardDeck)).
			init().
			SendMsg(windowSizeMsg).
			SendKeyRune(studyKey).
			SendKeyType(tea.KeyEnter).
			SendKeyRune(flashcard.ReviewScoreNormal.String()).
			SendKeyRune("?").
			Get()

		view := m.View()

		assert.NotContains(t, view, "q quit • ? more")
		assert.Contains(t, view, "q quit")
		assert.Contains(t, view, "? close help")
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
}
