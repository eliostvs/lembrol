package terminal_test

import (
	"container/list"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/eliostvs/remembercli/internal/flashcard"
	"github.com/eliostvs/remembercli/internal/terminal"
)

func TestMain(m *testing.M) {
	_ = os.Setenv("GLAMOUR_STYLE", "ascii")
	defer func() {
		_ = os.Unsetenv("GLAMOUR_STYLE")
	}()

	exitVal := m.Run()
	os.Exit(exitVal)
}

/*
 Test Utilities
*/

var breakLineMsg = tea.KeyMsg{Type: tea.KeyEnter, Alt: true}

var windowSizeMsg = tea.WindowSizeMsg{Width: 100, Height: 20}

func assertContainsMarkdown(t *testing.T, contains string, width int, content string) {
	t.Helper()

	content, _ = terminal.Markdown(width-4, content)
	content = strings.TrimSpace(content)
	assert.Contains(t, contains, content)
}

const (
	manyDecks      = "./testdata/many"
	fewDecks       = "./testdata/few"
	singleCardDeck = "./testdata/single"
	emptyDeck      = "./testdata/empty"
	noneDeck       = "./testdata/none"
	invalidDeck    = "./testdata/invalid"
	shortNamesDeck = "./testdata/short"
	longNamesDeck  = "./testdata/long"

	createKey    = "a"
	quitKey      = "q"
	studyKey     = "s"
	skipKey      = "s"
	deleteKey    = "x"
	renameKey    = "r"
	vimKeyDown   = "j"
	vimKeyUp     = "k"
	editKey      = "e"
	activePrompt = "│ "
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

	longestCard = flashcard.Card{
		Question: "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur",
		Answer:   "Maecenas condimentum neque nisl, eget pulvinar magna accumsan vitae. Quisque pretium nunc ipsum, volutpat tincidunt neque sagittis id. Phasellus ac dolor ac libero varius eleifend vel eu quam. Donec luctus suscipit ante vitae tincidunt. Praesent non purus blandit, molestie nisi id, gravida quam. Aliquam rutrum diam id libero fermentum dignissim",
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
	msgName := reflect.TypeOf(msg).Name()
	switch msgName {
	case "TickMsg", "blinkMsg":
		return true
	default:
		return false
	}
}

func (m *testModel) processCmd(cmd tea.Cmd) (msgs []tea.Msg) {
	if cmd == nil {
		return msgs
	}

	val := reflect.ValueOf(cmd())

	switch val.Kind() {
	case reflect.Slice:
		for i := 0; i < val.Len(); i++ {
			valCmd := val.Index(i).Interface().(tea.Cmd)
			msgs = append(msgs, m.processCmd(valCmd)...)
		}

	case reflect.Struct:
		msgs = append(msgs, val.Interface())
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
