package terminal_test

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dustin/go-humanize"
	"github.com/stretchr/testify/assert"

	"github.com/eliostvs/remembercli/internal/terminal"
	"github.com/eliostvs/remembercli/internal/test"
)

func TestCardsList(t *testing.T) {
	t.Run("shows deck with no cards", func(t *testing.T) {
		m, _ := newTestModel(emptyDeck).
			init().
			SendKeyType(tea.KeyEnter).
			Get()

		view := m.View()

		assert.Contains(t, view, "Empty")
		assert.Contains(t, view, "No items")
		assert.Contains(t, view, "a add • q quit • ? more ")
	})

	t.Run("truncates very long question text", func(t *testing.T) {
		m, _ := newTestModel(longNamesDeck).
			init().
			SendMsg(windowSizeMsg).
			SendKeyType(tea.KeyEnter).
			Get()

		view := m.View()

		assert.Contains(t, view, "Very Long Question & Answer")
		assert.Contains(t, view, "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididu…")
	})

	t.Run("shows deck with many cards", func(t *testing.T) {
		m, _ := newTestModel(fewDecks, terminal.WithClock(test.Clock{Time: oldestCard.ReviewedAt.Add(24 * time.Hour)})).
			init().
			SendMsg(windowSizeMsg).
			SendKeyType(tea.KeyEnter).
			Get()

		view := m.View()

		assert.Contains(t, view, "Golang A")
		assert.Contains(t, view, "6 items")
		assert.Contains(t, view, activePrompt+latestCard.Question)
		assert.Contains(t, view, fmt.Sprintf("%sLast review %s", activePrompt, humanize.Time(latestCard.ReviewedAt)))
		assert.Contains(t, view, secondLatestCard.Question)
		assert.Contains(t, view, fmt.Sprintf("Last review %s", humanize.Time(secondLatestCard.ReviewedAt)))
		assert.Contains(t, view, secondLatestCard.Question)
		assert.Contains(t, view, fmt.Sprintf("Last review %s", humanize.Time(secondLatestCard.ReviewedAt)))
		assert.Contains(t, view, "••")
		assert.Contains(t, view, "↑/k up • ↓/j down • / filter • a add • s study • q quit • ? more")
	})

	t.Run("closes the deck show navigates back to home page", func(t *testing.T) {
		m, _ := newTestModel(singleCardDeck).
			init().
			SendKeyType(tea.KeyEnter).
			SendKeyType(tea.KeyEsc).
			Get()

		assert.Contains(t, m.View(), "Decks")
	})

	t.Run("navigates around", func(t *testing.T) {
		tests := []struct {
			name string
			keys []string
			want string
		}{
			{
				name: "moves up, down using arrow keys",
				keys: []string{tea.KeyDown.String(), tea.KeyUp.String()},
				want: activePrompt + latestCard.Question,
			},
			{
				name: "moves down, up, up",
				keys: []string{tea.KeyDown.String(), vimKeyDown, vimKeyDown},
				want: activePrompt + "Question D",
			},
			{
				name: "moves right, left using arrow keys",
				keys: []string{tea.KeyRight.String(), tea.KeyLeft.String()},
				want: activePrompt + latestCard.Question,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var batch []tea.Msg
				for _, key := range tt.keys {
					batch = append(batch, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
				}

				m, _ := newTestModel(manyDecks).
					init().
					SendMsg(windowSizeMsg).
					SendKeyType(tea.KeyEnter).
					SendBatch(batch).
					Get()

				assert.Contains(t, m.View(), tt.want)
			})
		}
	})

	t.Run("starts study redirects to question page", func(t *testing.T) {
		m, _ := newTestModel(singleCardDeck).
			init().
			SendKeyType(tea.KeyEnter).
			SendKeyRune(studyKey).
			Get()

		assert.Contains(t, m.View(), "Question")
	})

	t.Run("ignores start review when deck has no due cards", func(t *testing.T) {
		m, _ := newTestModel(singleCardDeck, terminal.WithClock(test.Clock{Time: latestCard.ReviewedAt})).
			init().
			SendKeyType(tea.KeyEnter).
			SendKeyRune(studyKey).
			Get()

		assert.Contains(t, m.View(), "Golang One")
	})
}

func TestCardCreate(t *testing.T) {
	t.Run("shows create card page", func(t *testing.T) {
		m, _ := newTestModel(singleCardDeck).
			init().
			SendKeyType(tea.KeyEnter).
			SendKeyRune(createKey).
			Get()

		view := m.View()

		assert.Contains(t, view, "Golang One")
		assert.Contains(t, view, "Front")
		assert.Contains(t, view, "nter a question")
		assert.Contains(t, view, "Back")
		assert.Contains(t, view, "Enter an answer")
		assert.Contains(t, view, "↓ down • ↑ up • enter confirm • esc cancel")
	})

	t.Run("goes to deck page when card creation is canceled", func(t *testing.T) {
		m, _ := newTestModel(manyDecks).
			init().
			SendKeyType(tea.KeyEnter).
			SendKeyRune(createKey).
			SendKeyType(tea.KeyEsc).
			Get()

		view := m.View()

		assert.NotContains(t, view, "Front")
		assert.Contains(t, view, activePrompt+latestCard.Question)
	})

	t.Run("goes to deck page when card is created", func(t *testing.T) {
		m, _ := newTestModel(test.TempDirCopy(t, emptyDeck)).
			init().
			SendKeyType(tea.KeyEnter).
			SendKeyRune(createKey).
			SendKeyRune("New").
			SendKeyType(tea.KeyDown).
			SendKeyRune("New").
			SendKeyType(tea.KeyShiftTab).
			SendKeyRune(" question").
			SendKeyType(tea.KeyDown).
			SendKeyRune(" answer").
			SendKeyType(tea.KeyUp).
			SendKeyRune("?").
			SendKeyType(tea.KeyEnter).
			SendKeyType(tea.KeyEnter).
			Get()

		view := m.View()

		assert.Contains(t, view, "Empty")
		assert.Contains(t, view, "1 item")
		assert.Contains(t, view, activePrompt+"New question")
		assert.Contains(t, view, activePrompt+"Last review just now")
	})

	t.Run("goes error page when card creation fails", func(t *testing.T) {
		location, cleanup := test.TempReadOnlyDirCopy(t, emptyDeck)
		t.Cleanup(cleanup)

		m, _ := newTestModel(location).
			init().
			SendKeyType(tea.KeyEnter).
			SendKeyRune(createKey).
			SendKeyRune("Answer").
			SendKeyType(tea.KeyDown).
			SendKeyRune("Question").
			SendKeyType(tea.KeyEnter).
			Get()

		assert.Contains(t, m.View(), "Error")
	})

	t.Run("validates card creation", func(t *testing.T) {
		t.Run("missing question", func(t *testing.T) {
			m, _ := newTestModel(test.TempDirCopy(t, emptyDeck)).
				init().
				SendKeyType(tea.KeyEnter).
				SendKeyRune(createKey).
				SendKeyType(tea.KeyEnter).
				SendKeyRune("Answer").
				SendKeyType(tea.KeyEnter).
				Get()

			view := m.View()

			assert.Contains(t, view, "Front")
		})

		t.Run("missing answer", func(t *testing.T) {
			m, _ := newTestModel(test.TempDirCopy(t, emptyDeck)).
				init().
				SendKeyType(tea.KeyEnter).
				SendKeyRune(createKey).
				SendKeyRune("Question").
				SendKeyType(tea.KeyEnter).
				SendKeyType(tea.KeyEnter).
				Get()

			view := m.View()

			assert.Contains(t, view, "Front")
			assert.Contains(t, view, "Back")
		})
	})
}

func TestCardEdit(t *testing.T) {
	t.Run("shows card edit page", func(t *testing.T) {
		m, _ := newTestModel(singleCardDeck).
			init().
			SendKeyType(tea.KeyEnter).
			SendKeyRune(editKey).
			Get()

		view := m.View()

		assert.Contains(t, view, "Golang One")
		assert.Contains(t, view, "Back")
		assert.Contains(t, view, "> "+latestCard.Question)
		assert.Contains(t, view, "Front")
		assert.Contains(t, view, "> "+latestCard.Answer)
		assert.Contains(t, view, "↓ down • ↑ up • enter confirm • esc cancel")
	})

	t.Run("goes to deck page when card edit is canceled", func(t *testing.T) {
		m, _ := newTestModel(manyDecks).
			init().
			SendKeyType(tea.KeyEnter).
			SendKeyType(tea.KeyRight).
			SendKeyRune(editKey).
			SendKeyType(tea.KeyEsc).
			Get()

		view := m.View()

		assert.NotContains(t, view, "Front")
	})

	t.Run("goes to error page when card edition fails", func(t *testing.T) {
		location, cleanup := test.TempReadOnlyDirCopy(t, singleCardDeck)
		t.Cleanup(cleanup)

		m, _ := newTestModel(location).
			init().
			SendKeyType(tea.KeyEnter).
			SendKeyRune(editKey).
			SendKeyType(tea.KeyEnter).
			SendKeyType(tea.KeyEnter).
			Get()

		assert.Contains(t, m.View(), "Error")
	})

	t.Run("goes to deck page when card edition succeed", func(t *testing.T) {
		m, _ := newTestModel(test.TempDirCopy(t, manyDecks)).
			init().
			SendKeyType(tea.KeyEnter).
			SendKeyRune(editKey).
			SendKeyRune("--").
			SendKeyType(tea.KeyDown).
			SendKeyRune("--").
			SendKeyType(tea.KeyUp).
			SendKeyType(tea.KeyBackspace).
			SendKeyType(tea.KeyTab).
			SendKeyType(tea.KeyBackspace).
			SendKeyType(tea.KeyShiftTab).
			SendKeyRune("q").
			SendKeyType(tea.KeyTab).
			SendKeyRune("q").
			SendKeyType(tea.KeyEnter).
			Get()

		view := m.View()

		assert.Contains(t, view, activePrompt+latestCard.Question+"-q")
	})

	t.Run("validates card edition", func(t *testing.T) {
		t.Run("missing question", func(t *testing.T) {
			m, _ := newTestModel(test.TempDirCopy(t, shortNamesDeck)).
				init().
				SendKeyType(tea.KeyEnter).
				SendKeyRune(editKey).
				SendKeyType(tea.KeyBackspace).
				SendKeyType(tea.KeyEnter).
				SendKeyType(tea.KeyEnter).
				Get()

			assert.Contains(t, m.View(), "Front")
		})

		t.Run("missing answer", func(t *testing.T) {
			m, _ := newTestModel(test.TempDirCopy(t, shortNamesDeck)).
				init().
				SendKeyType(tea.KeyEnter).
				SendKeyRune(editKey).
				SendKeyType(tea.KeyDown).
				SendKeyType(tea.KeyBackspace).
				SendKeyType(tea.KeyEnter).
				Get()

			assert.Contains(t, m.View(), "Front")
		})
	})

	t.Run("write multiline text", func(t *testing.T) {
		m, _ := newTestModel(test.TempDirCopy(t, emptyDeck)).
			init().
			SendKeyType(tea.KeyEnter).
			SendKeyRune(createKey).
			SendKeyRune("Question first line").
			SendMsg(breakLineMsg).
			SendKeyRune("Question second line").
			SendKeyType(tea.KeyDown).
			SendKeyRune("Answer first line").
			SendMsg(breakLineMsg).
			SendKeyRune("Answer second line").
			SendKeyType(tea.KeyEnter).
			Get()

		view := m.View()

		assert.Contains(t, view, "Question first line¬Question second line")
	})
}

func TestCardDelete(t *testing.T) {
	t.Run("shows delete message", func(t *testing.T) {
		m, _ := newTestModel(fewDecks).
			init().
			SendMsg(windowSizeMsg).
			SendKeyType(tea.KeyEnter).
			SendKeyRune(deleteKey).
			Get()

		view := m.View()

		assert.Contains(t, view, "Delete this card?")
		assert.Contains(t, view, activePrompt+latestCard.Question)
		assert.Contains(t, view, fmt.Sprintf("%sLast review %s", activePrompt, humanize.Time(latestCard.ReviewedAt)))
	})

	t.Run("goes to deck page the card deletion is canceled", func(t *testing.T) {
		m, _ := newTestModel(manyDecks).
			init().
			SendMsg(windowSizeMsg).
			SendKeyType(tea.KeyEnter).
			SendKeyRune(deleteKey).
			SendKeyType(tea.KeyEsc).
			Get()

		view := m.View()

		assert.NotContains(t, view, "Delete this card?")
		assert.Contains(t, view, activePrompt+latestCard.Question)
	})

	t.Run("ignores delete action when deck has no cards", func(t *testing.T) {
		m, _ := newTestModel(noneDeck).
			init().
			SendKeyType(tea.KeyEnter).
			SendKeyRune(deleteKey).
			Get()

		assert.NotContains(t, m.View(), "Delete this card?")
	})

	t.Run("goes to error page when card deletion fails", func(t *testing.T) {
		location, cleanup := test.TempReadOnlyDirCopy(t, singleCardDeck)
		t.Cleanup(cleanup)

		m, _ := newTestModel(location).
			init().
			SendKeyType(tea.KeyEnter).
			SendKeyRune(deleteKey).
			SendKeyType(tea.KeyEnter).
			Get()

		assert.Contains(t, m.View(), "Error")
	})

	t.Run("goes to deck page when card is deleted", func(t *testing.T) {
		m, _ := newTestModel(test.TempDirCopy(t, manyDecks)).
			init().
			SendMsg(windowSizeMsg).
			SendKeyType(tea.KeyEnter).
			SendKeyRune(deleteKey).
			SendKeyType(tea.KeyEnter).
			Get()

		view := m.View()

		assert.NotContains(t, view, "Delete this card?")
		assert.NotContains(t, view, latestCard.Question)
	})
}
