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
			Init().
			SendKeyType(tea.KeyEnter).
			Get()

		view := m.View()

		assert.Contains(t, view, "Empty")
		assert.Contains(t, view, "No items")
		assert.Contains(t, view, "a add • q quit • ? more ")
	})

	t.Run("shows full help with card page without cards", func(t *testing.T) {
		m, _ := newTestModel(emptyDeck).
			Init().
			SendMsg(windowSizeMsg).
			SendKeyType(tea.KeyEnter).
			SendKeyRune(helpKey).
			Get()

		view := m.View()

		assert.Contains(t, view, "a add    q quit")
		assert.Contains(t, view, "? close help")
	})

	t.Run("shows deck with many cards", func(t *testing.T) {
		m, _ := newTestModel(fewDecks, terminal.WithClock(test.Clock{Time: oldestCard.ReviewedAt.Add(24 * time.Hour)})).
			Init().
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

	t.Run("shows full help when there are many cards", func(t *testing.T) {
		m, _ := newTestModel(manyDecks).
			Init().
			SendMsg(windowSizeMsg).
			SendKeyType(tea.KeyEnter).
			SendKeyRune(helpKey).
			Get()

		view := m.View()

		assert.Contains(t, view, "↑/k      up             / filter    q quit")
		assert.Contains(t, view, "↓/j      down           a add       ? close help")
		assert.Contains(t, view, "→/l/pgdn next page      e edit")
		assert.Contains(t, view, " ←/h/pgup prev page      x delete")
		assert.Contains(t, view, "g/home   go to start    s study")
		assert.Contains(t, view, "G/end    go to end")
	})

	t.Run("truncates very long question text", func(t *testing.T) {
		m, _ := newTestModel(longNamesDeck).
			Init().
			SendMsg(windowSizeMsg).
			SendKeyType(tea.KeyEnter).
			Get()

		view := m.View()

		assert.Contains(t, view, "Very Long Question & Answer")
		assert.Contains(t, view, "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididu…")
	})

	t.Run("shows homepage when deck is closed", func(t *testing.T) {
		m, _ := newTestModel(singleCardDeck).
			Init().
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
				keys: []string{tea.KeyDown.String(), keyDown, keyDown},
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
					Init().
					SendMsg(windowSizeMsg).
					SendKeyType(tea.KeyEnter).
					SendBatch(batch).
					Get()

				assert.Contains(t, m.View(), tt.want)
			})
		}
	})

	t.Run("shows questions when review starts", func(t *testing.T) {
		m, _ := newTestModel(singleCardDeck).
			Init().
			SendKeyType(tea.KeyEnter).
			SendKeyRune(studyKey).
			Get()

		assert.Contains(t, m.View(), "Question")
	})

	t.Run("does not start review when it does not have due cards", func(t *testing.T) {
		m, _ := newTestModel(singleCardDeck, terminal.WithClock(test.Clock{Time: latestCard.ReviewedAt})).
			Init().
			SendKeyType(tea.KeyEnter).
			SendKeyRune(studyKey).
			Get()

		assert.Contains(t, m.View(), "Golang One")
	})

	t.Run("filters cards", func(t *testing.T) {
		m, _ := newTestModel(manyDecks).
			Init().
			SendMsg(windowSizeMsg).
			SendKeyType(tea.KeyEnter).
			SendKeyRune(filterKey).
			SendKeyRune("Question A").
			SendKeyType(tea.KeyEnter).
			Get()

		view := m.View()

		assert.Contains(t, view, "“Question …” 6 items")
	})
}

func TestCardCreate(t *testing.T) {
	t.Run("shows create card form", func(t *testing.T) {
		m, _ := newTestModel(singleCardDeck).
			Init().
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

	t.Run("shows deck when the creation is canceled", func(t *testing.T) {
		m, _ := newTestModel(manyDecks).
			Init().
			SendKeyType(tea.KeyEnter).
			SendKeyRune(createKey).
			SendKeyType(tea.KeyEsc).
			Get()

		view := m.View()

		assert.NotContains(t, view, "Front")
		assert.Contains(t, view, activePrompt+latestCard.Question)
	})

	t.Run("shows deck when card is created", func(t *testing.T) {
		m, _ := newTestModel(test.TempDirCopy(t, emptyDeck)).
			Init().
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

	t.Run("shows error when card the creation fail", func(t *testing.T) {
		location, cleanup := test.TempReadOnlyDirCopy(t, emptyDeck)
		t.Cleanup(cleanup)

		m, _ := newTestModel(location).
			Init().
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
				Init().
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
				Init().
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

	t.Run("write multiline text", func(t *testing.T) {
		m, _ := newTestModel(test.TempDirCopy(t, emptyDeck)).
			Init().
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

func TestCardEdit(t *testing.T) {
	t.Run("shows edit card form", func(t *testing.T) {
		m, _ := newTestModel(singleCardDeck).
			Init().
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

	t.Run("shows error when the edition fail", func(t *testing.T) {
		location, cleanup := test.TempReadOnlyDirCopy(t, singleCardDeck)
		t.Cleanup(cleanup)

		m, _ := newTestModel(location).
			Init().
			SendKeyType(tea.KeyEnter).
			SendKeyRune(editKey).
			SendKeyType(tea.KeyEnter).
			SendKeyType(tea.KeyEnter).
			Get()

		assert.Contains(t, m.View(), "Error")
	})

	t.Run("shows deck when card is changed", func(t *testing.T) {
		m, _ := newTestModel(test.TempDirCopy(t, manyDecks)).
			Init().
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

}

func TestCardDelete(t *testing.T) {
	t.Run("confirms card deletion", func(t *testing.T) {
		m, _ := newTestModel(fewDecks).
			Init().
			SendMsg(windowSizeMsg).
			SendKeyType(tea.KeyEnter).
			SendKeyRune(deleteKey).
			Get()

		view := m.View()

		assert.Contains(t, view, "Delete this card?")
		assert.Contains(t, view, activePrompt+latestCard.Question)
		assert.Contains(t, view, fmt.Sprintf("%sLast review %s", activePrompt, humanize.Time(latestCard.ReviewedAt)))
	})

	t.Run("shows deck when the card deletion is canceled", func(t *testing.T) {
		m, _ := newTestModel(manyDecks).
			Init().
			SendMsg(windowSizeMsg).
			SendKeyType(tea.KeyEnter).
			SendKeyRune(deleteKey).
			SendKeyType(tea.KeyEsc).
			Get()

		view := m.View()

		assert.NotContains(t, view, "Delete this card?")
		assert.Contains(t, view, activePrompt+latestCard.Question)
	})

	t.Run("ignores inputs besides cancel or enter", func(t *testing.T) {
		tests := []struct {
			name string
			key  string
		}{
			{
				name: "delete",
				key:  deleteKey,
			},
			{
				name: "create",
				key:  createKey,
			},
			{
				name: "study",
				key:  studyKey,
			},
			{
				name: "edit",
				key:  editKey,
			},
			{
				name: "help",
				key:  helpKey,
			},
			{
				name: "down",
				key:  "down",
			},
			{
				name: "up",
				key:  "up",
			},
			{
				name: "left",
				key:  "left",
			},
			{
				name: "right",
				key:  "right",
			},
			{
				name: "home",
				key:  "home",
			},
			{
				name: "down",
				key:  "down",
			},
			{
				name: "filter",
				key:  filterKey,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				m, _ := newTestModel(test.TempDirCopy(t, manyDecks)).
					Init().
					SendKeyType(tea.KeyEnter).
					SendKeyRune(deleteKey).
					SendKeyRune(tt.key).
					Get()

				view := m.View()

				assert.Contains(t, view, "Delete this card?")
			})
		}
	})

	t.Run("ignores delete action when deck has no cards", func(t *testing.T) {
		m, _ := newTestModel(noneDeck).
			Init().
			SendKeyType(tea.KeyEnter).
			SendKeyRune(deleteKey).
			Get()

		assert.NotContains(t, m.View(), "Delete this card?")
	})

	t.Run("shows error when card deletion fails", func(t *testing.T) {
		location, cleanup := test.TempReadOnlyDirCopy(t, singleCardDeck)
		t.Cleanup(cleanup)

		m, _ := newTestModel(location).
			Init().
			SendKeyType(tea.KeyEnter).
			SendKeyRune(deleteKey).
			SendKeyType(tea.KeyEnter).
			Get()

		assert.Contains(t, m.View(), "Error")
	})

	t.Run("shows deck when card is deleted", func(t *testing.T) {
		m, _ := newTestModel(test.TempDirCopy(t, manyDecks)).
			Init().
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
