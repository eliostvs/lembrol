package tui

import (
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/eliostvs/lembrol/internal/clock"
	"github.com/eliostvs/lembrol/internal/flashcard"
)

const (
	appName = "Lembrol"
)

// MODEL

// ModelOption configure the Model options.
type ModelOption func(*Model)

// WithClock initializes the model with the given clock.
func WithClock(clock clock.Clock) ModelOption {
	return func(m *Model) {
		m.clock = clock
	}
}

// WithWindowSize initializes the model with the given width and height.
func WithWindowSize(width, height int) ModelOption {
	return func(m *Model) {
		m.width, m.height = width, height
	}
}

// Repository wraps the file system operation
// to be easier and quicker run the tests.
type Repository interface {
	List() []flashcard.Deck
	Create(name string, cards []flashcard.Card) (flashcard.Deck, error)
	Save(flashcard.Deck) error
	Delete(flashcard.Deck) error
}

// WithRepository configure the terminal with an alternative repository.
func WithRepository(factory func(clock.Clock) (Repository, error)) ModelOption {
	return func(m *Model) {
		m.repositoryFactory = factory
	}
}

type createdRepositoryMsg struct {
	repository Repository
}

// NewModel creates a new model instance given a decks location.
func NewModel(path string, debug bool, opts ...ModelOption) Model {
	shared := Shared{
		clock:  clock.New(),
		styles: NewStyles(lipgloss.DefaultRenderer()),
		debug:  debug,
	}
	m := Model{
		page: newLoadingPage(shared, appName, "Loading..."),
		repositoryFactory: func(c clock.Clock) (Repository, error) {
			c.Sleep(time.Second)
			return flashcard.NewRepository(path, c)
		},
		Shared: shared,
	}

	for _, opt := range opts {
		opt(&m)
	}

	return m
}

type Shared struct {
	repository Repository
	clock      clock.Clock
	width      int
	height     int
	styles     *Styles
	debug      bool
}

func (s *Shared) Log(msg string) {
	if s.debug {
		log.Println(msg)
	}
}

type Model struct {
	repositoryFactory func(clock.Clock) (Repository, error)
	page              tea.Model
	Shared
}

// MESSAGES

func fail(err error) tea.Msg {
	return setErrorPageMsg{err}
}

type setErrorPageMsg struct {
	err error
}

func showDecks(index int) tea.Cmd {
	return func() tea.Msg {
		return setDecksPageMsg{index: index}
	}
}

type setDecksPageMsg struct {
	index int
}

func showCards(index int, deck flashcard.Deck) tea.Cmd {
	return func() tea.Msg {
		return setCardsPageMsg{deck: deck, index: index}
	}
}

type setCardsPageMsg struct {
	index int
	deck  flashcard.Deck
}

func showStats(index int, card flashcard.Card, deck flashcard.Deck) tea.Cmd {
	return func() tea.Msg {
		return setStatsPageMsg{deck: deck, card: card, cardIndex: index}
	}
}

type setStatsPageMsg struct {
	cardIndex int
	deck      flashcard.Deck
	card      flashcard.Card
}

func startReview(d flashcard.Deck) tea.Cmd {
	return func() tea.Msg {
		return setReviewPageMsg{d}
	}
}

type setReviewPageMsg struct {
	flashcard.Deck
}

type setQuitPageMsg struct{}

func quit() tea.Msg {
	return setQuitPageMsg{}
}

// UPDATE

func (m Model) Init() tea.Cmd {
	m.Log("app: init")

	return tea.Batch(
		func() tea.Msg {
			repo, err := m.repositoryFactory(m.clock)
			if err != nil {
				return fail(err)
			}
			return createdRepositoryMsg{repo}
		},
		m.page.Init(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.Log(fmt.Sprintf("app: %T", msg))

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case createdRepositoryMsg:
		m.repository = msg.repository
		m.page = newDeckPage(m.Shared, 0)
		return m, m.page.Init()

	case setDecksPageMsg:
		m.page = newDeckPage(m.Shared, 0)
		return m, m.page.Init()

	case setCardsPageMsg:
		m.page = newCardPage(m.Shared, msg.deck)
		return m, m.page.Init()

	case setStatsPageMsg:
		m.page = newStatsModel(m.Shared, msg)
		return m, m.page.Init()

	case setReviewPageMsg:
		m.page = newReviewPage(m.Shared, flashcard.NewReview(msg.Deck, m.clock))
		return m, m.page.Init()

	case setErrorPageMsg:
		m.page = newErrorPage(m.Shared, msg.err)
		return m, m.page.Init()

	case setQuitPageMsg:
		m.page = newQuitModel(m.Shared)
		return m, m.page.Init()
	}

	m.page, cmd = m.page.Update(msg)

	return m, cmd
}

// VIEW

func (m Model) View() string {
	m.Log("app: view")

	return m.page.View()
}
