package terminal_test

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eliostvs/remembercli/internal/flashcard"
	"github.com/eliostvs/remembercli/internal/test"
)

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
		assert.Contains(t, view, "0: again")
		assert.Contains(t, view, "1: hard")
		assert.Contains(t, view, "2: normal")
		assert.Contains(t, view, "3: easy")
		assert.Contains(t, view, "4: super easy")
		assert.Contains(t, view, "esc: close")
		assert.Contains(t, view, "q: quit")
	})

	t.Run("wraps long answers", func(t *testing.T) {
		width := 70
		m, _ := newTestModel(longNamesDeck).
			init().
			SendKeyRune(studyKey).
			SendMsg(tea.WindowSizeMsg{Width: width}).
			SendKeyType(tea.KeyEnter).
			Get()

		view := m.View()

		assert.Contains(t, view, "Very Long Question & Answer")
		assertContainsMarkdown(t, view, width, longestCard.Answer)
	})

	t.Run("goes to deck page when the review is canceled", func(t *testing.T) {
		m, _ := newTestModel(manyDecks).
			init().
			SendKeyRune(studyKey).
			SendKeyType(tea.KeyEnter).
			SendKeyType(tea.KeyEsc).
			Get()

		assert.Contains(t, m.View(), "Deck")
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

	t.Run("quits the app from answer page", func(t *testing.T) {
		m, _ := newTestModel(test.TempDirCopy(t, singleCardDeck)).
			init().
			SendKeyRune(studyKey).
			SendKeyType(tea.KeyEnter).
			SendKeyRune(quitKey).
			Get()

		assert.Contains(t, m.View(), "Thanks for using Remember CLI!")
	})
}
