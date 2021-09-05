package terminal_test

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/eliostvs/remembercli/internal/terminal"
	"github.com/eliostvs/remembercli/internal/test"
)

func TestDecks(t *testing.T) {
	t.Run("shows home page with no decks", func(t *testing.T) {
		m, _ := newTestModel(t.TempDir()).
			init().
			Get()

		view := m.View()

		assert.Contains(t, m.View(), "No deck files found. Press 'a' to create one.")
		assert.NotContains(t, view, "•\n")
		assert.NotContains(t, view, "j/k, ↑/↓: choose")
		assert.NotContains(t, view, "h/l ←/→: page")
		assert.NotContains(t, view, "enter: select")
		assert.NotContains(t, view, "s: study")
		assert.NotContains(t, view, "x: delete")
		assert.Contains(t, view, "q: quit")
	})

	t.Run("shows home page with one deck", func(t *testing.T) {
		m, _ := newTestModel(singleCardDeck).
			init().
			Get()

		view := m.View()

		assert.Contains(t, view, activePrompt+itemPrompt+"Golang One")
		assert.Contains(t, view, activePrompt+"1 card | 1 due")
		assert.Contains(t, view, "1 deck")
		assert.NotContains(t, view, "•\n")
		assert.NotContains(t, view, "j/k, ↑/↓: choose")
		assert.NotContains(t, view, "h/l ←/→: page")
		assert.Contains(t, view, "enter: select")
		assert.Contains(t, view, "x: delete")
		assert.Contains(t, view, "r: rename")
		assert.Contains(t, view, "s: study")
		assert.Contains(t, view, "q: quit")
	})

	t.Run("shows home page with many decks", func(t *testing.T) {
		m, _ := newTestModel(manyDecks, terminal.WithClock(test.NewClock(oldestCard.ReviewedAt.Add(24*time.Hour*4)))).
			init().
			Get()

		view := m.View()

		assert.Contains(t, view, "Decks")
		assert.Contains(t, view, "6 decks")
		assert.Contains(t, view, activePrompt+itemPrompt+"Golang A")
		assert.Contains(t, view, activePrompt+"6 cards | 3 due")
		assert.Contains(t, view, itemPrompt+"Golang B")
		assert.Contains(t, view, "3 cards | 0 due")
		assert.NotContains(t, view, activePrompt+itemPrompt+"Golang B")
		assert.Contains(t, view, "••\n")
		assert.Contains(t, view, "a: add")
		assert.Contains(t, view, "j/k, ↑/↓: choose")
		assert.Contains(t, view, "h/l ←/→: page")
		assert.Contains(t, view, "enter: select")
		assert.Contains(t, view, "s: study")
		assert.Contains(t, view, "x: delete")
		assert.Contains(t, view, "q: quit")
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
				want: activePrompt + itemPrompt + "Golang B",
			},
			{
				name: "move down, down, down using vim keys",
				keys: []string{vimKeyDown, vimKeyDown, vimKeyDown},
				want: activePrompt + itemPrompt + "Golang D",
			},
			{
				name: "move down, up, up using arrow keys",
				keys: []string{tea.KeyDown.String(), tea.KeyUp.String(), tea.KeyUp.String()},
				want: activePrompt + itemPrompt + "Golang A",
			},
			{
				name: "move down, up, up using vim key",
				keys: []string{vimKeyDown, vimKeyUp, vimKeyUp},
				want: activePrompt + itemPrompt + "Golang A",
			},
			{
				name: "move down, right using arrow keys",
				keys: []string{tea.KeyDown.String(), tea.KeyRight.String()},
				want: activePrompt + itemPrompt + "Golang F",
			},
			{
				name: "move right, down, left using vim keys",
				keys: []string{vimKeyRight, vimKeyDown, vimKeyLeft},
				want: activePrompt + itemPrompt + "Golang A",
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
					SendBatch(batch).
					Get()

				assert.Contains(t, m.View(), tt.want)
			})
		}
	})

	t.Run("quits the app from home page", func(t *testing.T) {
		m, _ := newTestModel(t.TempDir()).
			init().
			SendKeyRune(quitKey).
			Get()

		assert.Contains(t, m.View(), "Thanks for using Remember CLI!")
	})

	t.Run("ignores actions when there are no decks", func(t *testing.T) {
		tests := []struct {
			name string
			key  string
		}{
			{
				name: "selects deck",
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
				key:  vimKeyUp,
			},
			{
				name: "chooses deck bellow",
				key:  vimKeyDown,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				m, cmd := newTestModel(t.TempDir()).
					init().
					SendKeyRune(tt.key).
					Get()

				assert.Nil(t, cmd)
				assert.Contains(t, m.View(), "No deck files found")
			})
		}
	})

	t.Run("only cards with due cards can start a review", func(t *testing.T) {
		m, cmd := newTestModel(manyDecks, terminal.WithClock(test.Clock{Time: oldestCard.ReviewedAt})).
			init().
			SendKeyRune(studyKey).
			Get()

		assert.Nil(t, cmd)
		assert.Contains(t, m.View(), "Decks")
		assert.NotContains(t, m.View(), "Question")
	})

	t.Run("shows study page when study starts", func(t *testing.T) {
		m, _ := newTestModel(singleCardDeck).
			init().
			SendKeyRune(studyKey).
			Get()

		assert.Contains(t, m.View(), latestCard.Question)
	})

	t.Run("shows deck page when a deck is selected", func(t *testing.T) {
		m, _ := newTestModel(singleCardDeck).
			init().
			SendKeyType(tea.KeyEnter).
			Get()

		view := m.View()

		assert.Contains(t, view, "Deck")
		assert.NotContains(t, view, "Decks")
	})

	t.Run("always selects the first card in the deck", func(t *testing.T) {
		m, _ := newTestModel(manyDecks).
			init().
			SendKeyType(tea.KeyEnter).
			SendKeyRune(vimKeyDown).
			SendKeyType(tea.KeyDown).
			SendKeyType(tea.KeyEsc).
			SendKeyRune(vimKeyDown).
			SendKeyType(tea.KeyEnter).
			Get()

		assert.Contains(t, m.View(), activePrompt+itemPrompt+latestCard.Question)
	})
}

func TestDeckCreate(t *testing.T) {
	t.Run("shows create deck page", func(t *testing.T) {
		m, _ := newTestModel(noneDeck).
			init().
			SendKeyRune(createKey).
			Get()

		view := m.View()

		assert.Contains(t, view, "Add Deck")
		assert.Contains(t, view, "Choose a deck name")
		assert.Contains(t, view, "enter: create")
		assert.Contains(t, view, "esc: cancel")
	})

	t.Run("goes to home page when creation is canceled", func(t *testing.T) {
		m, _ := newTestModel(manyDecks).
			init().
			SendKeyType(tea.KeyRight).
			SendKeyRune(createKey).
			SendKeyType(tea.KeyEsc).
			Get()

		view := m.View()

		assert.Contains(t, view, "6 decks")
		assert.Contains(t, view, "Golang F")
	})

	t.Run("validates deck creation", func(t *testing.T) {
		m, _ := newTestModel(noneDeck).
			init().
			SendKeyRune(createKey).
			SendKeyType(tea.KeyEnter).
			Get()

		assert.Contains(t, m.View(), "Choose a deck name")
	})

	t.Run("goes to deck page when deck is created", func(t *testing.T) {
		m, _ := newTestModel(t.TempDir()).
			init().
			SendKeyRune(createKey).
			SendText("Golang q").
			SendKeyType(tea.KeyEnter).
			Get()

		view := m.View()

		assert.Contains(t, view, "Golang q")
		assert.NotContains(t, view, "Your currently have not cards")
	})

	t.Run("goes to error page when deck creation fails", func(t *testing.T) {
		location, cleanup := test.TempReadOnlyDirCopy(t, t.TempDir())
		t.Cleanup(cleanup)

		m, _ := newTestModel(location).
			init().
			SendKeyRune(createKey).
			SendText("Golang New.toml").
			SendKeyType(tea.KeyEnter).
			Get()

		assert.Contains(t, m.View(), "Error")
	})
}

func TestDeckRename(t *testing.T) {
	t.Run("shows rename page with few pages", func(t *testing.T) {
		m, _ := newTestModel(fewDecks).
			init().
			SendKeyRune(renameKey).
			Get()

		view := m.View()

		assert.Contains(t, view, "Rename this deck?")
		assert.NotContains(t, view, "•\n")
		assert.Contains(t, view, "Name:")
		assert.Contains(t, view, "enter: rename")
		assert.Contains(t, view, "esc: cancel")
	})

	t.Run("shows rename page with many pages", func(t *testing.T) {
		m, _ := newTestModel(manyDecks).
			init().
			SendKeyRune(renameKey).
			Get()

		view := m.View()

		assert.Contains(t, view, "Rename this deck?")
		assert.Contains(t, view, "Name:")
		assert.Contains(t, view, "••\n")
		assert.Contains(t, view, "enter: rename")
		assert.Contains(t, view, "esc: cancel")
	})

	t.Run("goes to home page when rename is canceled", func(t *testing.T) {
		m, _ := newTestModel(manyDecks).
			init().
			SendKeyType(tea.KeyRight).
			SendKeyRune(renameKey).
			SendKeyType(tea.KeyEsc).
			Get()

		view := m.View()

		assert.NotContains(t, view, "Rename this deck?")
		assert.Contains(t, view, "Golang F")
	})

	t.Run("validates deck rename", func(t *testing.T) {
		m, _ := newTestModel(test.TempDirCopy(t, shortNamesDeck)).
			init().
			SendKeyRune(renameKey).
			SendKeyType(tea.KeyBackspace).
			SendKeyType(tea.KeyBackspace).
			SendKeyType(tea.KeyEnter).
			Get()

		assert.Contains(t, m.View(), "Rename this deck?")
	})

	t.Run("goes to home page when deck is renamed", func(t *testing.T) {
		m, _ := newTestModel(test.TempDirCopy(t, manyDecks)).
			init().
			SendKeyType(tea.KeyRight).
			SendKeyRune(renameKey).
			SendKeyType(tea.KeyBackspace).
			SendKeyRune("q").
			SendKeyType(tea.KeyEnter).
			Get()

		view := m.View()

		assert.Contains(t, view, "Decks")
		assert.NotContains(t, view, "Rename this deck?")
		assert.Contains(t, view, activePrompt+itemPrompt+"Golang q")
	})

	t.Run("goes to error page when deck rename fails", func(t *testing.T) {
		location, cleanup := test.TempReadOnlyDirCopy(t, singleCardDeck)
		t.Cleanup(cleanup)

		m, _ := newTestModel(location).
			init().
			SendKeyRune(renameKey).
			SendText(" Change").
			SendKeyType(tea.KeyEnter).
			Get()

		assert.Contains(t, m.View(), "Error")
	})
}

func TestDeckDelete(t *testing.T) {
	t.Run("shows delete deck page with few decks", func(t *testing.T) {
		m, _ := newTestModel(test.TempDirCopy(t, fewDecks)).
			init().
			SendKeyRune(deleteKey).
			Get()

		view := m.View()

		assert.Contains(t, view, "Decks")
		assert.Contains(t, view, "Delete this deck?")
		assert.Contains(t, view, activePrompt+itemPrompt+"Golang A")
		assert.Contains(t, view, activePrompt+"6 cards | 6 due")
		assert.Contains(t, view, itemPrompt+"Golang B")
		assert.Contains(t, view, "2 cards | 2 due")
		assert.NotContains(t, view, "•\n")
		assert.Contains(t, view, "enter: delete")
		assert.Contains(t, view, "esc: cancel")
		assert.Contains(t, view, "q: quit")
	})

	t.Run("shows delete deck page with many decks", func(t *testing.T) {
		m, _ := newTestModel(test.TempDirCopy(t, manyDecks)).
			init().
			SendKeyRune(deleteKey).
			Get()

		view := m.View()

		assert.Contains(t, view, "Decks")
		assert.Contains(t, view, "Delete this deck?")
		assert.Contains(t, view, activePrompt+itemPrompt+"Golang A")
		assert.Contains(t, view, activePrompt+"6 cards | 6 due")
		assert.Contains(t, view, itemPrompt+"Golang B")
		assert.Contains(t, view, "3 cards | 3 due")
		assert.Contains(t, view, "••\n")
		assert.Contains(t, view, "enter: delete")
		assert.Contains(t, view, "esc: cancel")
		assert.Contains(t, view, "q: quit")
	})

	t.Run("goes to home page when deck deletion is canceled", func(t *testing.T) {
		m, _ := newTestModel(manyDecks).
			init().
			SendKeyType(tea.KeyRight).
			SendKeyRune(deleteKey).
			SendKeyType(tea.KeyEsc).
			Get()

		view := m.View()

		assert.NotContains(t, view, "Delete this deck?")
		assert.Contains(t, view, "Golang F")
	})

	t.Run("quits the app from delete page", func(t *testing.T) {
		m, _ := newTestModel(singleCardDeck).
			init().
			SendKeyRune(deleteKey).
			SendKeyRune(quitKey).
			Get()

		assert.Contains(t, m.View(), "Thanks for using Remember CLI!")
	})

	t.Run("goes to home page when deck is deleted", func(t *testing.T) {
		m, _ := newTestModel(test.TempDirCopy(t, manyDecks)).
			init().
			SendKeyType(tea.KeyRight).
			SendKeyType(tea.KeyDown).
			SendKeyType(tea.KeyDown).
			SendKeyType(tea.KeyDown).
			SendKeyRune(deleteKey).
			SendKeyType(tea.KeyEnter).
			Get()

		view := m.View()

		assert.Contains(t, view, "5 deck")
		assert.Contains(t, view, "Golang A")
		assert.NotContains(t, view, "Golang F")
		assert.NotContains(t, view, "h/l ←/→: page")
		assert.NotContains(t, view, "•\n")
	})

	t.Run("goes to error page when deletion fails", func(t *testing.T) {
		location, cleanup := test.TempReadOnlyDirCopy(t, singleCardDeck)
		t.Cleanup(cleanup)

		m, _ := newTestModel(location).
			init().
			SendKeyRune(deleteKey).
			SendKeyType(tea.KeyEnter).
			Get()

		assert.Contains(t, m.View(), "Error")
	})
}
