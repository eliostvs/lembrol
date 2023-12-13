package terminal_test

import (
	"container/list"
	"errors"
	"fmt"
	"github.com/eliostvs/lembrol/internal/test"
	"os"
	"reflect"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/eliostvs/lembrol/internal/clock"
	"github.com/eliostvs/lembrol/internal/flashcard"
	"github.com/eliostvs/lembrol/internal/terminal"
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

const (
	testWidth      = 100
	testHeight     = 20
	manyDecks      = "./testdata/many"
	fewDecks       = "./testdata/few"
	singleCardDeck = "./testdata/single"
	errorDeck      = "./testdata/error"
	invalidDeck    = "./testdata/invalid"
	emptyDeck      = "./testdata/empty"
	noneDeck       = "./testdata/none"
	longNamesDeck  = "./testdata/long"
	errorDeckName  = "Error"

	createKey    = "a"
	saveKey      = "ctrl+s"
	quitKey      = "q"
	studyKey     = "s"
	skipKey      = "s"
	deleteKey    = "x"
	keyDown      = "j"
	keyUp        = "k"
	editKey      = "e"
	helpKey      = "?"
	filterKey    = "/"
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

func newTestRepository(t *testing.T, p string, c clock.Clock) (*repository, error) {
	r, err := flashcard.NewRepository(test.TempCopyDir(t, p), c)
	if err != nil {
		return nil, err
	}

	return &repository{r}, nil
}

type repository struct {
	*flashcard.Repository
}

func (r *repository) Create(name string, cards []flashcard.Card) (flashcard.Deck, error) {
	if name == errorDeckName {
		return flashcard.Deck{}, errors.New("failed")
	}

	return r.Repository.Create(name, cards)
}

func (r *repository) Save(deck flashcard.Deck) error {
	if deck.Name == errorDeckName {
		return errors.New("failed")
	}

	return r.Repository.Save(deck)
}

func (r *repository) Delete(deck flashcard.Deck) error {
	if deck.Name == errorDeckName {
		return errors.New("failed")
	}

	return r.Repository.Delete(deck)
}

func newTestModel(t *testing.T, path string, opts ...terminal.ModelOption) *testModel {
	t.Helper()

	defaultOpts := []terminal.ModelOption{
		terminal.WithWindowSize(testWidth, testHeight),
		terminal.WithRepository(
			func(clock clock.Clock) (terminal.Repository, error) {
				return newTestRepository(t, path, clock)
			},
		),
	}
	return &testModel{
		model: terminal.NewModel(path, append(defaultOpts, opts...)...),
		queue: newMsgQueue(),
		ignoreMsg: map[string]struct{}{
			"TickMsg":         {},
			"BlinkMsg":        {},
			"blinkMsg":        {},
			"initialBlinkMsg": {},
		},
		observer: func(tea.Model) {
		},
	}
}

type testModel struct {
	model     tea.Model
	cmd       tea.Cmd
	queue     msgQueue
	ignoreMsg map[string]struct{}
	observer  func(tea.Model)
}

func (m *testModel) Init() *testModel {
	return m.processMsg(m.processCmd(m.model.Init()))
}

func (m *testModel) processMsg(msgs []tea.Msg) *testModel {
	for _, msg := range msgs {
		m.queue.Enqueue(msg)
	}

	for !m.queue.Empty() {
		msg := m.queue.Dequeue()

		if m.shouldSkip(msg) {
			continue
		}

		m.model, m.cmd = m.model.Update(msg)
		m.observer(m.model)

		if m.cmd != nil {
			for _, msg := range m.processCmd(m.cmd) {
				m.queue.Enqueue(msg)
			}
		}
	}

	return m
}

func (m *testModel) shouldSkip(msg tea.Msg) bool {
	name := reflect.TypeOf(msg).Name()
	_, ok := m.ignoreMsg[name]
	return ok
}

func (m *testModel) processCmd(cmd tea.Cmd) (msgs []tea.Msg) {
	if cmd == nil {
		return msgs
	}

	val := reflect.ValueOf(cmd())

	switch val.Kind() {
	case reflect.Slice:
		for i := 0; i < val.Len(); i++ {
			valCmd, ok := val.Index(i).Interface().(tea.Cmd)
			if ok {
				msgs = append(msgs, m.processCmd(valCmd)...)
			} else {
				msgs = append(msgs, val.Interface())
			}
		}
	default:
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

func (m *testModel) Get() tea.Model {
	return m.model
}

func (m *testModel) Peek(fn func(tea.Model)) *testModel {
	fn(m.model)
	return m
}

func (m *testModel) Print() *testModel {
	return m.Peek(func(m tea.Model) { fmt.Println(m.View()) })
}

func (m *testModel) WithObserver(fn func(tea.Model)) *testModel {
	m.observer = fn
	return m
}

func (m *testModel) ForceUpdate(msg tea.Msg) *testModel {
	m.model, m.cmd = m.model.Update(msg)
	return m
}
