package terminal_test

import (
	"container/list"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dustin/go-humanize"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eliostvs/remembercli/internal/flashcard"
	"github.com/eliostvs/remembercli/internal/terminal"
	"github.com/eliostvs/remembercli/internal/test"
)

const (
	manyDecks      = "./testdata/many"
	fewDecks       = "./testdata/few"
	singleCardDeck = "./testdata/single"
	emptyDeck      = "./testdata/empty"
	noneDeck       = "./testdata/none"
	invalidDeck    = "./testdata/invalid"
	shortNamesDeck = "./testdata/short"
	longNamesDeck  = "./testdata/long"

	createKey   = "a"
	quitKey     = "q"
	studyKey    = "s"
	skipKey     = "s"
	deleteKey   = "x"
	renameKey   = "r"
	vimKeyDown  = "j"
	vimKeyLeft  = "h"
	vimKeyRight = "l"
	vimKeyUp    = "k"
	editKey     = "e"

	activePrompt = "│ "
	itemPrompt   = "• "
)

var (
	latestCard = flashcard.Card{
		Question:   "Question A",
		Answer:     "Answer A",
		ReviewedAt: time.Date(2021, 1, 8, 15, 4, 0, 0, time.UTC),
	}

	secondLatestCard = flashcard.Card{
		Question:   "Question B",
		Answer:     "Answer B",
		ReviewedAt: time.Date(2021, 1, 6, 15, 4, 0, 0, time.UTC),
	}

	oldestCard = flashcard.Card{
		Question:   "Question F",
		Answer:     "Answer F",
		ReviewedAt: time.Date(2021, 1, 2, 15, 4, 0, 0, time.UTC),
	}
)

func newMsgQueue() msgQueue {
	return msgQueue{list.New()}
}

type msgQueue struct {
	data *list.List
}

func (q msgQueue) Enqueue(msg tea.Msg) {
	q.data.PushBack(msg)
}

func (q msgQueue) Dequeue() tea.Msg {
	elem := q.data.Front()

	if elem == nil {
		panic("queue is empty!")
	}

	q.data.Remove(elem)
	return elem.Value
}

func (q msgQueue) Empty() bool {
	return q.data.Len() == 0
}

func newTestModel(location string, opts ...terminal.ModelOption) *testModel {
	return &testModel{
		m: terminal.NewModel(location, append(opts, terminal.WithInitialDelay(0))...),
		q: newMsgQueue(),
	}
}

type testModel struct {
	m tea.Model
	c tea.Cmd
	q msgQueue
}

// the tea.Batch used in the Model.init method returns a private value, batchMsg, that is a slice of Cmd,
// so we are using the reflect package to iterate over it.
func (m *testModel) init() *testModel {
	return m.processMsg(m.processCmd(m.m.Init()))
}

func (m *testModel) processMsg(msgs []tea.Msg) *testModel {
	for _, msg := range msgs {
		m.q.Enqueue(msg)
	}

	for !m.q.Empty() {
		msg := m.q.Dequeue()
		if m.skip(msg) {
			continue
		}

		m.m, m.c = m.m.Update(msg)
		if m.c != nil {
			for _, msg := range m.processCmd(m.c) {
				m.q.Enqueue(msg)
			}
		}
	}

	return m
}

func (m *testModel) skip(msg tea.Msg) bool {
	_, ok := msg.(spinner.TickMsg)
	return ok
}

func (m *testModel) processCmd(cmd tea.Cmd) []tea.Msg {
	val := reflect.ValueOf(cmd())

	if val.Kind() != reflect.Slice {
		return []tea.Msg{val.Interface()}
	}

	var msgs []tea.Msg
	for i := 0; i < val.Len(); i++ {
		msgs = append(msgs, val.Index(i).Call(nil)[0].Interface())
	}

	return msgs
}

func (m *testModel) SendMsg(msg tea.Msg) *testModel {
	return m.processMsg([]tea.Msg{msg})
}

func (m *testModel) SendBatch(msgs []tea.Msg) *testModel {
	return m.processMsg(msgs)
}

func (m *testModel) SendKeyRune(r string) *testModel {
	return m.SendMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(r)})
}

func (m *testModel) SendKeyType(t tea.KeyType) *testModel {
	return m.SendMsg(tea.KeyMsg{Type: t})
}

func (m *testModel) SendText(text string) *testModel {
	for _, char := range text {
		m = m.SendKeyRune(string(char))
	}
	return m
}

func (m *testModel) Get() (tea.Model, tea.Cmd) {
	return m.m, m.c
}

func (m *testModel) Peek(fn func(tea.Model)) *testModel {
	fn(m.m)
	return m
}

func (m *testModel) Print() *testModel {
	return m.Peek(func(m tea.Model) { fmt.Println(m.View()) })
}

func (m *testModel) ForceUpdate(msg tea.Msg) *testModel {
	m.m, m.c = m.m.Update(msg)
	return m
}

// Loading

func TestLoading(t *testing.T) {
	t.Run("goes to home page when loading is successful", func(t *testing.T) {
		m, _ := newTestModel(manyDecks).
			init().
			Get()

		assert.Contains(t, m.View(), "Decks")
	})

	t.Run("goes to error page when loading fails", func(t *testing.T) {
		m, _ := newTestModel(invalidDeck).
			init().
			Get()

		assert.Contains(t, m.View(), "Error")
	})

	t.Run("shows loading page", func(t *testing.T) {
		newTestModel(manyDecks).
			Peek(func(m tea.Model) {
				assert.Contains(t, m.View(), "Remember")
				assert.Contains(t, m.View(), "⣾  Loading...")
			}).
			ForceUpdate(spinner.TickMsg{}).
			Peek(func(m tea.Model) {
				assert.Contains(t, m.View(), "⣽  Loading...")
			}).
			ForceUpdate(spinner.TickMsg{}).
			Peek(func(m tea.Model) {
				assert.Contains(t, m.View(), "⣻  Loading...")
			}).
			ForceUpdate(spinner.TickMsg{}).
			Peek(func(m tea.Model) {
				assert.Contains(t, m.View(), "⢿  Loading...")
			})
	})
}

// Error

func TestError(t *testing.T) {
	t.Run("shows error page", func(t *testing.T) {
		m, _ := newTestModel(invalidDeck).
			init().
			Get()

		view := m.View()

		assert.Contains(t, view, "Error")
		assert.Contains(t, view, "unmarshall deck")
	})
}

// Decks

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

		assert.Contains(t, m.View(), "Thanks for using RememberCLI!")
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

		assert.Contains(t, m.View(), "Thanks for using RememberCLI!")
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

// Cards

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

		assert.Contains(t, m.View(), "Thanks for using RememberCLI!")
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

		assert.Contains(t, m.View(), "Thanks for using RememberCLI!")
	})
}

// Question

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

	t.Run("wraps to long answers", func(t *testing.T) {
		m, _ := newTestModel(longNamesDeck).
			init().
			SendKeyRune(studyKey).
			SendMsg(tea.WindowSizeMsg{Width: 70}).
			Get()

		view := m.View()

		assert.Equal(t, 6, strings.Count(view, activePrompt))
		assert.Contains(t, view, "Very Long Question & Answer")
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

		assert.Contains(t, m.View(), "Thanks for using RememberCLI!")
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
}

// Answer

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

	t.Run("wraps to long answers", func(t *testing.T) {
		m, _ := newTestModel(longNamesDeck).
			init().
			SendKeyRune(studyKey).
			SendMsg(tea.WindowSizeMsg{Width: 70}).
			SendKeyType(tea.KeyEnter).
			Get()

		view := m.View()

		assert.Equal(t, 6, strings.Count(view, activePrompt))
		assert.Contains(t, view, "Very Long Question & Answer")
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

		assert.Contains(t, m.View(), "Thanks for using RememberCLI!")
	})
}

// Review

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

		assert.Contains(t, m.View(), "Thanks for using RememberCLI")
	})
}
