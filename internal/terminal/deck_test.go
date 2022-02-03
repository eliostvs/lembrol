package terminal_test

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/eliostvs/lembrol/internal/terminal"
	"github.com/eliostvs/lembrol/internal/test"
)

func TestDecksList(t *testing.T) {
	t.Parallel()

	t.Run("shows homepage without decks", func(t *testing.T) {
		m, _ := newTestModel(t.TempDir()).
			Init().
			Get()

		view := m.View()

		assert.NotContains(t, view, "↑/k up")
		assert.NotContains(t, view, "↓/j down")
		assert.NotContains(t, view, "/ filter")
		assert.NotContains(t, view, "enter open")
		assert.Contains(t, view, "a add • q quit • ? more")
	})

	t.Run("shows fulls help when there are no decks", func(t *testing.T) {
		m, _ := newTestModel(t.TempDir()).
			Init().
			SendMsg(windowSizeMsg).
			SendKeyRune(helpKey).
			Get()

		view := m.View()

		assert.Contains(t, view, "a add    q quit")
		assert.Contains(t, view, "? close help")
	})

	t.Run("shows homepage with many decks", func(t *testing.T) {
		m, _ := newTestModel(manyDecks, terminal.WithClock(test.NewClock(oldestCard.ReviewedAt.Add(24*time.Hour*4)))).
			Init().
			SendMsg(windowSizeMsg).
			Get()

		view := m.View()

		assert.Contains(t, view, "Decks")
		assert.Contains(t, view, "6 items")
		assert.Contains(t, view, activePrompt+"Golang A")
		assert.Contains(t, view, activePrompt+"6 cards | 3 due")
		assert.Contains(t, view, "Golang B")
		assert.Contains(t, view, "3 cards | 0 due")
		assert.NotContains(t, view, activePrompt+"Golang B")
		assert.Contains(t, view, "••")
		assert.Contains(t, view, "↑/k up • ↓/j down • / filter • a add • enter open • q quit • ? more")
	})

	t.Run("shows full help when there are many decks", func(t *testing.T) {
		m, _ := newTestModel(manyDecks).
			Init().
			SendMsg(windowSizeMsg).
			SendKeyRune(helpKey).
			Get()

		view := m.View()

		assert.Contains(t, view, "↑/k      up             /     filter    q quit")
		assert.Contains(t, view, "↓/j      down           a     add       ? close help")
		assert.Contains(t, view, "→/l/pgdn next page      enter open")
		assert.Contains(t, view, "g/home   go to start    x     delete")
		assert.Contains(t, view, "G/end    go to end      s     study")
	})

	t.Run("navigates around", func(t *testing.T) {
		tests := []struct {
			name string
			keys []string
			want string
		}{
			{
				name: "move down using arrows keys",
				keys: []string{tea.KeyDown.String()},
				want: activePrompt + "Golang B",
			},
			{
				name: "move down, down, down using vim keys",
				keys: []string{keyDown, keyDown, keyDown},
				want: activePrompt + "Golang D",
			},
			{
				name: "move down, up, up using arrow keys",
				keys: []string{tea.KeyDown.String(), tea.KeyUp.String(), tea.KeyUp.String()},
				want: activePrompt + "Golang A",
			},
			{
				name: "move down, up, up using vim key",
				keys: []string{keyDown, keyUp, keyUp},
				want: activePrompt + "Golang A",
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
					SendBatch(batch).
					Get()

				assert.Contains(t, m.View(), tt.want)
			})
		}
	})

	t.Run("quits the app", func(t *testing.T) {
		m, _ := newTestModel(t.TempDir()).
			Init().
			SendKeyRune(quitKey).
			Get()

		assert.Contains(t, m.View(), "Thanks for using Lembrol!")
	})

	t.Run("ignores actions when there are no decks", func(t *testing.T) {
		tests := []struct {
			name string
			key  string
		}{
			{
				name: "open deck",
				key:  tea.KeyEnter.String(),
			},
			{
				name: "starts review",
				key:  studyKey,
			},
			{
				name: "deletes deck",
				key:  studyKey,
			},
			{
				name: "chooses deck above",
				key:  keyUp,
			},
			{
				name: "chooses deck bellow",
				key:  keyDown,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				m, _ := newTestModel(t.TempDir()).
					Init().
					SendKeyRune(tt.key).
					Get()

				assert.Contains(t, m.View(), "No items")
			})
		}
	})

	t.Run("does not start review when it does not have due cards", func(t *testing.T) {
		m, _ := newTestModel(manyDecks, terminal.WithClock(test.Clock{Time: oldestCard.ReviewedAt})).
			Init().
			SendKeyRune(studyKey).
			Get()

		assert.Contains(t, m.View(), "Decks")
		assert.NotContains(t, m.View(), "Question")
	})

	t.Run("shows questions when review starts", func(t *testing.T) {
		m, _ := newTestModel(singleCardDeck).
			Init().
			SendKeyRune(studyKey).
			Get()

		assert.Contains(t, m.View(), latestCard.Question)
	})

	t.Run("shows deck when it is open", func(t *testing.T) {
		m, _ := newTestModel(singleCardDeck).
			Init().
			SendKeyType(tea.KeyEnter).
			Get()

		view := m.View()

		assert.Contains(t, view, "Golang One")
		assert.NotContains(t, view, "Decks")
	})

	t.Run("filters decks", func(t *testing.T) {
		m, _ := newTestModel(manyDecks).
			Init().
			SendMsg(windowSizeMsg).
			SendKeyRune(filterKey).
			SendKeyRune("Golang A").
			SendKeyType(tea.KeyEnter).
			Get()

		view := m.View()

		assert.Contains(t, view, "“Golang A” 6 items")
	})
}

func TestDeckCreate(t *testing.T) {
	t.Parallel()

	t.Run("shows crate form", func(t *testing.T) {
		m, _ := newTestModel(noneDeck).
			Init().
			SendKeyRune(createKey).
			Get()

		view := m.View()

		assert.Contains(t, view, "New Deck")
		assert.Contains(t, view, "enter confirm • esc cancel")
	})

	t.Run("shows homepage when the creation is canceled", func(t *testing.T) {
		m, _ := newTestModel(manyDecks).
			Init().
			SendKeyRune(createKey).
			SendKeyType(tea.KeyEsc).
			Get()

		view := m.View()

		assert.Contains(t, view, "Decks")
		assert.NotContains(t, view, "Add Deck")
	})

	t.Run("validates deck name can't be empty", func(t *testing.T) {
		m, _ := newTestModel(noneDeck).
			Init().
			SendKeyRune(createKey).
			SendKeyType(tea.KeyEnter).
			Get()

		assert.Contains(t, m.View(), "New Deck")
	})

	t.Run("validates deck name can't have multi line", func(t *testing.T) {
		m, _ := newTestModel(noneDeck).
			Init().
			SendKeyRune(createKey).
			SendKeyRune("First Line").
			SendMsg(breakLineMsg).
			SendKeyRune("Second Line").
			Get()

		assert.Contains(t, m.View(), "First LineSecond Line")
	})

	t.Run("shows homepage when deck is created", func(t *testing.T) {
		m, _ := newTestModel(t.TempDir()).
			Init().
			SendKeyRune(createKey).
			SendKeyRune("Golang q").
			SendKeyType(tea.KeyEnter).
			Get()

		view := m.View()

		assert.Contains(t, view, "Decks")
		assert.Contains(t, view, "1 item")
	})

	t.Run("shows error when create deck fail", func(t *testing.T) {
		location, cleanup := test.TempReadOnlyDir(t)
		t.Cleanup(cleanup)

		m, _ := newTestModel(location).
			Init().
			SendKeyRune(createKey).
			SendKeyRune("Golang New.toml").
			SendKeyType(tea.KeyEnter).
			Get()

		assert.Contains(t, m.View(), "Error")
	})
}

func TestDeckEdit(t *testing.T) {
	t.Parallel()

	t.Run("shows edit form", func(t *testing.T) {
		m, _ := newTestModel(fewDecks).
			Init().
			SendMsg(windowSizeMsg).
			SendKeyRune(editKey).
			Get()

		view := m.View()

		assert.Contains(t, view, "Edit Deck")
		assert.Contains(t, view, "enter confirm • esc cancel")
	})

	t.Run("shows homepage when deck is edited", func(t *testing.T) {
		m, _ := newTestModel(test.TempDirCopy(t, manyDecks)).
			Init().
			SendMsg(windowSizeMsg).
			SendKeyRune(editKey).
			SendKeyType(tea.KeyBackspace).
			SendKeyRune("Q").
			SendKeyType(tea.KeyEnter).
			Get()

		view := m.View()

		assert.Contains(t, view, "Decks")
		assert.NotContains(t, view, "Rename this deck?")
		assert.Contains(t, view, activePrompt+"Golang Q")
	})

	t.Run("shows error when edition fail", func(t *testing.T) {
		location, cleanup := test.TempReadOnlyDirCopy(t, singleCardDeck)
		t.Cleanup(cleanup)

		m, _ := newTestModel(location).
			Init().
			SendKeyRune(editKey).
			SendKeyRune(" Change").
			SendKeyType(tea.KeyEnter).
			Get()

		assert.Contains(t, m.View(), "Error")
	})
}

func TestDeckDelete(t *testing.T) {
	t.Parallel()

	t.Run("confirms the deletion", func(t *testing.T) {
		m, _ := newTestModel(test.TempDirCopy(t, fewDecks)).
			Init().
			SendMsg(windowSizeMsg).
			SendKeyRune(deleteKey).
			Get()

		view := m.View()

		assert.Contains(t, view, "Delete this deck?")
		assert.Contains(t, view, activePrompt+"Golang A")
		assert.Contains(t, view, activePrompt+"6 cards | 6 due")
		assert.Contains(t, view, "enter confirm • q quit")
	})

	t.Run("shows homepage when the deletion is canceled", func(t *testing.T) {
		m, _ := newTestModel(manyDecks).
			Init().
			SendKeyRune(deleteKey).
			SendKeyType(tea.KeyEsc).
			Get()

		view := m.View()

		assert.NotContains(t, view, "Delete this deck?")
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
					SendKeyRune(deleteKey).
					SendKeyRune(tt.key).
					Get()

				view := m.View()

				assert.Contains(t, view, "Delete this deck?")
			})
		}
	})

	t.Run("shows homepage when the deck is deleted", func(t *testing.T) {
		m, _ := newTestModel(test.TempDirCopy(t, manyDecks)).
			Init().
			SendKeyRune(deleteKey).
			SendKeyType(tea.KeyEnter).
			Get()

		view := m.View()

		assert.Contains(t, view, "5 items")
		assert.Contains(t, view, "↑/k up • ↓/j down • / filter • a add • enter open • q quit • ? more")
	})

	t.Run("shows error when the deletion fail", func(t *testing.T) {
		location, cleanup := test.TempReadOnlyDirCopy(t, singleCardDeck)
		t.Cleanup(cleanup)

		m, _ := newTestModel(location).
			Init().
			SendKeyRune(deleteKey).
			SendKeyType(tea.KeyEnter).
			Get()

		assert.Contains(t, m.View(), "Error")
	})
}
