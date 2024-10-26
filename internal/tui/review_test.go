package tui_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/eliostvs/lembrol/internal/flashcard"
	"github.com/eliostvs/lembrol/internal/tui"
)

func TestQuestion(t *testing.T) {
	t.Parallel()

	t.Run(
		"shows question", func(t *testing.T) {
			view := newTestModel(t, singleCardDeck).
				Init().
				SendKeyRune(studyKey).
				Get().
				View()

			assert.Contains(t, view, "Question A")
			assert.Contains(t, view, "Golang One")
			assert.Contains(t, view, "1 of 1")
			assert.Contains(t, view, "enter answer • q quit")
			assert.NotContains(t, view, "s skip")
		},
	)

	t.Run(
		"does not show skip in a single card deck", func(t *testing.T) {
			view := newTestModel(t, singleCardDeck).
				Init().
				SendKeyRune(studyKey).
				Get().
				View()

			assert.Contains(t, view, "enter answer • q quit")
			assert.NotContains(t, view, "s skip")
		},
	)

	t.Run(
		"shows skip option in a multiple card deck", func(t *testing.T) {
			view := newTestModel(t, fewDecks).
				Init().
				SendKeyRune(studyKey).
				Get().
				View()

			assert.Contains(t, view, "s skip • enter answer • q quit")
		},
	)

	t.Run(
		"goes to deck page when the review is canceled", func(t *testing.T) {
			view := newTestModel(t, manyDecks).
				Init().
				SendKeyRune(studyKey).
				SendKeyType(tea.KeyEsc).
				Get().
				View()

			assert.Contains(t, view, "6 items")
		},
	)

	t.Run(
		"goes to answer page", func(t *testing.T) {
			view := newTestModel(t, singleCardDeck).
				Init().
				SendKeyRune(studyKey).
				SendKeyType(tea.KeyEnter).
				Get().
				View()

			assert.Contains(t, view, "Answer")
		},
	)

	t.Run(
		"skips the current card", func(t *testing.T) {
			questionA := "Question A"
			questionB := "Question B"
			var nextQuestion string

			choose := func(s string) string {
				if strings.Contains(s, questionA) {
					return questionB
				}
				return questionA
			}

			newTestModel(t, fewDecks).
				Init().
				SendKeyType(tea.KeyDown).
				SendKeyRune(studyKey).
				Peek(
					func(m tea.Model) {
						assert.Contains(t, m.View(), "1 of 2")
						nextQuestion = choose(m.View())
					},
				).
				SendKeyRune(skipKey).
				Peek(
					func(m tea.Model) {
						view := m.View()
						assert.Contains(t, view, "1 of 2")
						assert.Contains(t, view, nextQuestion)
						nextQuestion = choose(m.View())
					},
				).
				SendKeyRune(skipKey).
				Peek(
					func(m tea.Model) {
						view := m.View()
						assert.Contains(t, view, "1 of 2")
						assert.Contains(t, view, nextQuestion)
					},
				)
		},
	)

	t.Run(
		"does not show skip option in the last question", func(t *testing.T) {
			view := newTestModel(t, fewDecks).
				Init().
				SendKeyType(tea.KeyDown).
				SendKeyRune(studyKey).
				SendKeyType(tea.KeyEnter).
				SendKeyRune(flashcard.ReviewScoreNormal.String()).
				Get().
				View()

			assert.Contains(t, view, "2 of 2")
			assert.NotContains(t, view, "skip")
		},
	)

	t.Run(
		"changes the layout when the window resize", func(t *testing.T) {
			var before string

			after := newTestModel(t, fewDecks, tui.WithWindowSize(testWidth, testHeight*2)).
				Init().
				SendKeyType(tea.KeyDown).
				SendKeyRune(studyKey).
				SendKeyType(tea.KeyEnter).
				SendKeyRune(flashcard.ReviewScoreNormal.String()).
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

func TestAnswer(t *testing.T) {
	t.Parallel()

	t.Run(
		"shows answer page", func(t *testing.T) {
			view := newTestModel(t, singleCardDeck).
				Init().
				SendKeyRune(studyKey).
				SendKeyType(tea.KeyEnter).
				Get().
				View()

			assert.Contains(t, view, "Answer")
			assert.Contains(t, view, "Golang One")
			assert.Contains(t, view, "1 of 1")
			assert.Contains(t, view, latestCard.Answer)
			assert.Contains(t, view, "1 hard • 2 normal • 3 easy • 4 very easy • q quit • ? more")
		},
	)

	t.Run(
		"shows full help with many cards", func(t *testing.T) {
			view := newTestModel(t, manyDecks).
				Init().
				SendKeyRune(studyKey).
				SendKeyType(tea.KeyEnter).
				SendKeyRune(helpKey).
				Get().
				View()

			assert.Contains(t, view, "0 again    1 hard         q quit")
			assert.Contains(t, view, "2 normal       ? close help")
			assert.Contains(t, view, "3 easy")
			assert.Contains(t, view, "4 very easy")
		},
	)

	t.Run(
		"shows full help with single deck", func(t *testing.T) {
			view := newTestModel(t, singleCardDeck).
				Init().
				SendKeyRune(studyKey).
				SendKeyType(tea.KeyEnter).
				SendKeyRune(helpKey).
				Get().
				View()

			assert.NotContains(t, view, "0 again")
			assert.Contains(t, view, "1 hard         q quit")
			assert.Contains(t, view, "2 normal       ? close help")
			assert.Contains(t, view, "3 easy")
			assert.Contains(t, view, "4 very easy")
		},
	)

	t.Run(
		"goes to deck page when the review is canceled", func(t *testing.T) {
			view := newTestModel(t, manyDecks).
				Init().
				SendKeyRune(studyKey).
				SendKeyType(tea.KeyEnter).
				SendKeyType(tea.KeyEsc).
				Get().
				View()

			assert.Contains(t, view, "6 items")
		},
	)

	t.Run(
		"goes to next review card when the card rating is successful", func(t *testing.T) {
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
				t.Run(
					tt.name, func(t *testing.T) {
						view := newTestModel(t, fewDecks).
							Init().
							SendKeyRune(studyKey).
							SendKeyType(tea.KeyEnter).
							SendKeyRune(tt.args.String()).
							Get().
							View()

						assert.Contains(t, view, tt.want)
					},
				)
			}
		},
	)

	t.Run(
		"goes to error page when the rating fails", func(t *testing.T) {
			view := newTestModel(t, errorDeck).
				Init().
				SendKeyRune(studyKey).
				SendKeyType(tea.KeyEnter).
				SendKeyRune(flashcard.ReviewScoreNormal.String()).
				Get().
				View()

			assert.Contains(t, view, "Error")
		},
	)

	t.Run(
		"does nothing when rate values are out of valid range", func(t *testing.T) {
			tests := []struct {
				name  string
				score string
			}{
				{
					name:  "less than zero",
					score: "-1",
				},
				{
					name:  "more than 4",
					score: "5",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name, func(t *testing.T) {
						view := newTestModel(t, singleCardDeck).
							Init().
							SendKeyRune(studyKey).
							SendKeyType(tea.KeyEnter).
							SendKeyRune(tt.score).
							Get().
							View()

						assert.Contains(t, view, "Answer A")
					},
				)
			}
		},
	)

	t.Run(
		"goes to review page when the review ends", func(t *testing.T) {
			var showLoading bool

			view := newTestModel(t, singleCardDeck).
				WithObserver(
					func(m tea.Model) {
						if strings.Contains(m.View(), "Scoring card...") {
							showLoading = true
						}
					},
				).
				Init().
				SendKeyRune(studyKey).
				SendKeyType(tea.KeyEnter).
				SendKeyRune(flashcard.ReviewScoreNormal.String()).
				Get().
				View()

			assert.Contains(t, view, "Congratulations")
			assert.True(t, showLoading)
		},
	)

	t.Run(
		"changes the layout when the window resize", func(t *testing.T) {
			var before string

			after := newTestModel(t, fewDecks, tui.WithWindowSize(testWidth, testHeight*2)).
				Init().
				SendKeyType(tea.KeyDown).
				SendKeyRune(studyKey).
				SendKeyType(tea.KeyEnter).
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

func TestReview(t *testing.T) {
	t.Parallel()

	t.Run(
		"shows review page", func(t *testing.T) {
			view := newTestModel(t, singleCardDeck).
				Init().
				SendKeyRune(studyKey).
				SendKeyType(tea.KeyEnter).
				SendKeyRune(flashcard.ReviewScoreNormal.String()).
				Get().
				View()

			assert.Contains(t, view, "Congratulations!")
			assert.Contains(t, view, "1 card reviewed")
			assert.Contains(t, view, "q quit")
		},
	)

	t.Run(
		"goes to home page when review ends", func(t *testing.T) {
			view := newTestModel(t, singleCardDeck).
				Init().
				SendKeyRune(studyKey).
				SendKeyType(tea.KeyEnter).
				SendKeyRune(flashcard.ReviewScoreNormal.String()).
				SendKeyType(tea.KeyEsc).
				Get().
				View()

			assert.Contains(t, view, "Decks")
		},
	)

	t.Run(
		"changes the layout when the window resize", func(t *testing.T) {
			var before string

			after := newTestModel(t, singleCardDeck, tui.WithWindowSize(testWidth, testHeight*2)).
				Init().
				SendKeyRune(studyKey).
				SendKeyType(tea.KeyEnter).
				SendKeyRune(flashcard.ReviewScoreNormal.String()).
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
