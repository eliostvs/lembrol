package tui_test

import (
	"container/list"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/eliostvs/lembrol/internal/test"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/eliostvs/lembrol/internal/clock"
	"github.com/eliostvs/lembrol/internal/flashcard"
	"github.com/eliostvs/lembrol/internal/tui"
)

func TestMain(m *testing.M) {
	_ = os.Setenv("GLAMOUR_STYLE", "ascii")
	_ = os.Setenv("NO_COLOR", "1")
	defer func() {
		_ = os.Unsetenv("GLAMOUR_STYLE")
		_ = os.Unsetenv("NO_COLOR")
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
	cancelKey    = "ctrl+c"
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
		LastReview: time.Date(2021, 1, 8, 15, 0, 0, 0, time.UTC),
	}

	secondLatestCard = flashcard.Card{
		Question:   "Question B",
		Answer:     "Answer B",
		LastReview: time.Date(2021, 1, 6, 15, 0, 0, 0, time.UTC),
	}

	oldestCard = flashcard.Card{
		Question:   "Question F",
		Answer:     "Answer F",
		LastReview: time.Date(2021, 1, 2, 15, 0, 0, 0, time.UTC),
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

func newTestModel(t *testing.T, path string, opts ...tui.ModelOption) *testModel {
	t.Helper()

	defaultOpts := []tui.ModelOption{
		tui.WithWindowSize(testWidth, testHeight),
		tui.WithRepository(
			func(clock clock.Clock) (tui.Repository, error) {
				return newTestRepository(t, path, clock)
			},
		),
	}
	return &testModel{
		model: tui.NewModel(path, false, append(defaultOpts, opts...)...),
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
	return m.SendMsg(keyPressMsg(r))
}

func (m *testModel) SendKeyType(t rune) *testModel {
	return m.SendMsg(tea.KeyPressMsg{Code: t})
}

type renderedModel struct {
	model tea.Model
}

func (m renderedModel) View() string {
	return ansi.Strip(m.model.View().Content)
}

func (m renderedModel) Raw() tea.Model {
	return m.model
}

func viewText(m tea.Model) string {
	return ansi.Strip(m.View().Content)
}

func (m *testModel) Get() renderedModel {
	return renderedModel{model: m.model}
}

func (m *testModel) Peek(fn func(tea.Model)) *testModel {
	fn(m.model)
	return m
}

func (m *testModel) Print() *testModel {
	return m.Peek(func(m tea.Model) { fmt.Println(viewText(m)) })
}

func (m *testModel) WithObserver(fn func(tea.Model)) *testModel {
	m.observer = fn
	return m
}

func (m *testModel) ForceUpdate(msg tea.Msg) *testModel {
	m.model, m.cmd = m.model.Update(msg)
	return m
}

func keyPressMsg(input string) tea.KeyPressMsg {
	switch strings.ToLower(input) {
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "left":
		return tea.KeyPressMsg{Code: tea.KeyLeft}
	case "right":
		return tea.KeyPressMsg{Code: tea.KeyRight}
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "esc", "escape":
		return tea.KeyPressMsg{Code: tea.KeyEsc}
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	case "shift+tab":
		return tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	}

	runes := []rune(input)
	if len(runes) == 1 {
		return tea.KeyPressMsg{Text: input, Code: runes[0]}
	}

	return tea.KeyPressMsg{Text: input, Code: tea.KeyExtended}
}
