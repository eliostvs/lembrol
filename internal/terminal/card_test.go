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

func TestCards(t *testing.T) {
	t.Run("shows deck with no cards", func(t *testing.T) {
		m, _ := newTestModel(emptyDeck).
			init().
			SendKeyType(tea.KeyEnter).
			Get()

		view := m.View()

		assert.Contains(t, view, "Empty")
		assert.Contains(t, view, "0 card | 0 due")
		assert.Contains(t, view, "Your currently have not cards. Press 'a' to create one.")
		assert.NotContains(t, view, "•\n")
		assert.Contains(t, view, "a: add")
		assert.NotContains(t, view, "j/k, ↑/↓: choose")
		assert.NotContains(t, view, "s: study")
		assert.NotContains(t, view, "e: edit")
		assert.NotContains(t, view, "x: delete")
		assert.Contains(t, view, "esc: close")
		assert.Contains(t, view, "q: quit")
	})

	t.Run("truncates very long question text", func(t *testing.T) {
		m, _ := newTestModel(longNamesDeck).
			init().
			SendKeyType(tea.KeyEnter).
			SendMsg(tea.WindowSizeMsg{Width: 69}).
			Get()

		view := m.View()

		assert.Contains(t, view, "Very Long Question & Answer")
		assert.Contains(t, view, "Lorem ipsum dolor sit amet, consectetur adipiscing elit…")
	})

	t.Run("shows deck with one card", func(t *testing.T) {
		m, _ := newTestModel(singleCardDeck).
			init().
			SendKeyType(tea.KeyEnter).
			Get()

		view := m.View()

		assert.Contains(t, view, "Golang One")
		assert.Contains(t, view, "1 card | 1 due")
		assert.Contains(t, view, "a: add")
		assert.NotContains(t, view, "j/k, ↑/↓: choose")
		assert.NotContains(t, view, "h/l ←/→: page")
		assert.Contains(t, view, "e: edit")
		assert.Contains(t, view, "x: delete")
		assert.Contains(t, view, "esc: close")
		assert.Contains(t, view, "q: quit")
	})

	t.Run("shows deck with many cards", func(t *testing.T) {
		m, _ := newTestModel(fewDecks, terminal.WithClock(test.Clock{Time: oldestCard.ReviewedAt.Add(24 * time.Hour)})).
			init().
			SendKeyType(tea.KeyEnter).
			Get()

		view := m.View()

		assert.Contains(t, view, "Deck")
		assert.Contains(t, view, "Golang A")
		assert.Contains(t, view, "6 cards | 1 due")
		assert.Contains(t, view, activePrompt+itemPrompt+latestCard.Question)
		assert.Contains(t, view, fmt.Sprintf("%sLast review %s", activePrompt, humanize.Time(latestCard.ReviewedAt)))
		assert.Contains(t, view, itemPrompt+secondLatestCard.Question)
		assert.Contains(t, view, fmt.Sprintf("Last review %s", humanize.Time(secondLatestCard.ReviewedAt)))
		assert.Contains(t, view, itemPrompt+secondLatestCard.Question)
		assert.Contains(t, view, fmt.Sprintf("Last review %s", humanize.Time(secondLatestCard.ReviewedAt)))
		assert.Contains(t, view, "••\n")
		assert.Contains(t, view, "a: add")
		assert.Contains(t, view, "j/k, ↑/↓: choose")
		assert.Contains(t, view, "h/l ←/→: page")
		assert.Contains(t, view, "s: study")
		assert.Contains(t, view, "e: edit")
		assert.Contains(t, view, "x: delete")
		assert.Contains(t, view, "esc: close")
		assert.Contains(t, view, "q: quit")
	})

	t.Run("closes the deck show navigates back to decks list page", func(t *testing.T) {
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
				want: activePrompt + itemPrompt + latestCard.Question,
			},
			{
				name: "moves up, down using vim keys",
				keys: []string{vimKeyDown, vimKeyUp},
				want: activePrompt + itemPrompt + latestCard.Question,
			},
			{
				name: "moves down, up, up",
				keys: []string{tea.KeyDown.String(), vimKeyDown, vimKeyDown},
				want: activePrompt + itemPrompt + "Question D",
			},
			{
				name: "moves right using arrow keys",
				keys: []string{tea.KeyRight.String()},
				want: activePrompt + itemPrompt + "Question F",
			},
			{
				name: "moves right using vim keys",
				keys: []string{vimKeyRight},
				want: activePrompt + itemPrompt + "Question F",
			},
			{
				name: "moves right, left using arrow keys",
				keys: []string{tea.KeyRight.String(), tea.KeyLeft.String()},
				want: activePrompt + itemPrompt + latestCard.Question,
			},
			{
				name: "moves right, left using vim keys",
				keys: []string{vimKeyRight, vimKeyLeft},
				want: activePrompt + itemPrompt + latestCard.Question,
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
		m, cmd := newTestModel(singleCardDeck, terminal.WithClock(test.Clock{Time: latestCard.ReviewedAt})).
			init().
			SendKeyType(tea.KeyEnter).
			SendKeyRune(studyKey).
			Get()

		assert.Nil(t, cmd)
		assert.Contains(t, m.View(), "Deck")
	})

	t.Run("quits the app from deck page", func(t *testing.T) {
		m, _ := newTestModel(singleCardDeck).
			init().
			SendKeyType(tea.KeyEnter).
			SendKeyRune(quitKey).
			Get()

		assert.Contains(t, m.View(), "Thanks for using Remember CLI!")
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

		assert.Contains(t, view, "Add Card")
		assert.Contains(t, view, "Back")
		assert.Contains(t, view, "Front")
		assert.Contains(t, view, "nter a question")
		assert.Contains(t, view, "nter an answer")
		assert.Contains(t, view, "tab, shift+tab, ↑/↓: field")
		assert.Contains(t, view, "enter: confirm")
		assert.Contains(t, view, "esc: cancel")
	})

	t.Run("goes to deck page when card creation is canceled", func(t *testing.T) {
		m, _ := newTestModel(manyDecks).
			init().
			SendKeyType(tea.KeyEnter).
			SendKeyType(tea.KeyRight).
			SendKeyRune(createKey).
			SendKeyType(tea.KeyEsc).
			Get()

		view := m.View()

		assert.NotContains(t, view, "Add Card")
		assert.Contains(t, view, activePrompt+itemPrompt+oldestCard.Question)
	})

	t.Run("goes to deck page when card is created", func(t *testing.T) {
		m, _ := newTestModel(test.TempDirCopy(t, emptyDeck)).
			init().
			SendKeyType(tea.KeyEnter).
			SendKeyRune(createKey).
			SendText("New").
			SendKeyType(tea.KeyDown).
			SendText("New").
			SendKeyType(tea.KeyShiftTab).
			SendText(" question").
			SendKeyType(tea.KeyDown).
			SendText(" answer").
			SendKeyType(tea.KeyUp).
			SendKeyRune("?").
			SendKeyType(tea.KeyEnter).
			SendKeyType(tea.KeyEnter).
			Get()

		view := m.View()

		assert.Contains(t, view, "Deck")
		assert.Contains(t, view, "Empty")
		assert.Contains(t, view, "1 card | 1 due")
		assert.Contains(t, view, activePrompt+itemPrompt+"New question")
		assert.Contains(t, view, activePrompt+"Last review just now")
	})

	t.Run("goes error page when card creation fails", func(t *testing.T) {
		location, cleanup := test.TempReadOnlyDirCopy(t, emptyDeck)
		t.Cleanup(cleanup)

		m, _ := newTestModel(location).
			init().
			SendKeyType(tea.KeyEnter).
			SendKeyRune(createKey).
			SendText("Answer").
			SendKeyType(tea.KeyDown).
			SendText("Question").
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
				SendText("Answer").
				SendKeyType(tea.KeyEnter).
				Get()

			view := m.View()

			assert.Contains(t, view, "Add Card")
			assert.NotContains(t, view, "Your currently have not cards. Press 'a' to create one.")
		})

		t.Run("missing answer", func(t *testing.T) {
			m, _ := newTestModel(test.TempDirCopy(t, emptyDeck)).
				init().
				SendKeyType(tea.KeyEnter).
				SendKeyRune(createKey).
				SendText("Question").
				SendKeyType(tea.KeyEnter).
				SendKeyType(tea.KeyEnter).
				Get()

			view := m.View()

			assert.Contains(t, view, "Add Card")
			assert.NotContains(t, view, "Your currently have not cards. Press 'a' to create one.")
		})
	})
}

func TestCardEdit(t *testing.T) {
	t.Run("shows card edit page", func(t *testing.T) {
		m, _ := newTestModel(singleCardDeck).
			init().
			SendKeyType(tea.KeyEnter).
			SendKeyRune("e").
			Get()

		view := m.View()

		assert.Contains(t, view, "Edit Card")
		assert.Contains(t, view, "Back")
		assert.Contains(t, view, "Front")
		assert.Contains(t, view, latestCard.Question[1:])
		assert.Contains(t, view, latestCard.Answer)
		assert.Contains(t, view, "tab, shift+tab, ↑/↓: field")
		assert.Contains(t, view, "enter: confirm")
		assert.Contains(t, view, "esc: cancel")
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

		assert.NotContains(t, view, "Edit Card")
		assert.Contains(t, view, activePrompt+itemPrompt+oldestCard.Question)
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
			SendKeyType(tea.KeyRight).
			SendKeyRune(editKey).
			SendText("--").
			SendKeyType(tea.KeyDown).
			SendText("--").
			SendKeyType(tea.KeyUp).
			SendKeyType(tea.KeyBackspace).
			SendKeyType(tea.KeyTab).
			SendKeyType(tea.KeyBackspace).
			SendKeyType(tea.KeyShiftTab).
			SendText("q").
			SendKeyType(tea.KeyTab).
			SendText("q").
			SendKeyType(tea.KeyEnter).
			Get()

		view := m.View()

		assert.NotContains(t, view, "Edit Card")
		assert.Contains(t, view, activePrompt+itemPrompt+oldestCard.Question+"-q")
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

			assert.Contains(t, m.View(), "Edit Card")
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

			assert.Contains(t, m.View(), "Edit Card")
		})
	})
}

func TestCardDelete(t *testing.T) {
	t.Run("shows delete card page with few cards", func(t *testing.T) {
		m, _ := newTestModel(fewDecks).
			init().
			SendKeyType(tea.KeyEnter).
			SendKeyRune(deleteKey).
			Get()

		view := m.View()

		assert.Contains(t, view, "Deck")
		assert.Contains(t, view, "Delete this card?")
		assert.Contains(t, view, activePrompt+itemPrompt+latestCard.Question)
		assert.Contains(t, view, fmt.Sprintf("%sLast review %s", activePrompt, humanize.Time(latestCard.ReviewedAt)))
		assert.Contains(t, view, itemPrompt+secondLatestCard.Question)
		assert.Contains(t, view, fmt.Sprintf("Last review %s", humanize.Time(secondLatestCard.ReviewedAt)))
		assert.Contains(t, view, "enter: delete")
		assert.Contains(t, view, "esc: cancel")
		assert.Contains(t, view, "q: quit")
	})

	t.Run("truncates very long question text", func(t *testing.T) {
		m, _ := newTestModel(longNamesDeck).
			init().
			SendKeyType(tea.KeyEnter).
			SendKeyRune(deleteKey).
			SendMsg(tea.WindowSizeMsg{Width: 69}).
			Get()

		view := m.View()

		assert.Contains(t, view, "Delete this card?")
		assert.Contains(t, view, "Lorem ipsum dolor sit amet, consectetur adipiscing elit…")
	})

	t.Run("goes to deck page the card deletion is canceled", func(t *testing.T) {
		m, _ := newTestModel(manyDecks).
			init().
			SendKeyType(tea.KeyEnter).
			SendKeyType(tea.KeyRight).
			SendKeyRune(deleteKey).
			SendKeyType(tea.KeyEsc).
			Get()

		view := m.View()

		assert.NotContains(t, view, "Delete this card?")
		assert.Contains(t, view, activePrompt+itemPrompt+oldestCard.Question)
	})

	t.Run("ignores delete action when deck has no cards", func(t *testing.T) {
		m, _ := newTestModel(noneDeck).
			init().
			SendKeyType(tea.KeyEnter).
			SendKeyRune(deleteKey).
			SendKeyType(tea.KeyEsc).
			Get()

		assert.Contains(t, m.View(), "No deck files found.")
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
			SendKeyType(tea.KeyEnter).
			SendKeyType(tea.KeyRight).
			SendKeyType(tea.KeyDown).
			SendKeyType(tea.KeyDown).
			SendKeyRune(deleteKey).
			SendKeyType(tea.KeyEnter).
			Get()

		view := m.View()

		assert.Contains(t, view, "Golang A")
		assert.Contains(t, view, "5 cards | 5 due")
		assert.Contains(t, view, latestCard.Question)
		assert.NotContains(t, view, oldestCard.Question)
	})

	t.Run("quits the app from delete page", func(t *testing.T) {
		m, _ := newTestModel(singleCardDeck).
			init().
			SendKeyType(tea.KeyEnter).
			SendKeyRune(deleteKey).
			SendKeyRune(quitKey).
			Get()

		assert.Contains(t, m.View(), "Thanks for using Remember CLI!")
	})
}
