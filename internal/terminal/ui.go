package terminal

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/eliostvs/remembercli/internal/flashcard"
)

const (
	initialDelay = time.Millisecond * 800
)

type ModelOption func(*Model)

func WithInitialDelay(delay time.Duration) ModelOption {
	return func(m *Model) {
		m.initialDelay = delay
	}
}

func WithClock(clock flashcard.Clock) ModelOption {
	return func(m *Model) {
		m.clock = clock
	}
}

// MODEL

// NewModel creates a new model instance given a decks location.
func NewModel(location string, opts ...ModelOption) Model {
	spin := spinner.NewModel()
	spin.Spinner = spinner.Dot

	m := Model{
		Spinner:      spin,
		clock:        flashcard.NewClock(),
		initialDelay: initialDelay,
		location:     location,
		templates:    newTemplates(),
	}

	for _, opt := range opts {
		opt(&m)
	}

	return m
}

type Model struct {
	Error   string
	Spinner spinner.Model

	clock        flashcard.Clock
	initialDelay time.Duration
	location     string
	repository   *flashcard.Repository
	templates    *templates
	width        int

	page        page
	cardModel   cardModel
	deckModel   deckModel
	reviewModel reviewModel
}

// UPDATE

func createRepository(location string, clock flashcard.Clock) tea.Msg {
	repo, err := flashcard.NewRepository(location, clock)
	if err != nil {
		return failed(err)
	}
	return createdRepositoryMsg{repo}
}

type createdRepositoryMsg struct {
	*flashcard.Repository
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.Tick(m.initialDelay, func(time.Time) tea.Msg {
		return createRepository(m.location, m.clock)
	}), spinner.Tick)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = Width(msg)

	case spinner.TickMsg:
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd

	case createdRepositoryMsg:
		m.repository = msg.Repository
		m.deckModel = newDeckModel(m.repository.List(), m.repository)
		m.page = Decks
		return m, nil

	case setDecksPageMsg:
		m.page = Decks
		return m, nil

	case setCardsPageMsg:
		m.cardModel = newCardModel(msg.Deck, m.clock, m.repository)
		m.page = Cards
		return m, nil

	case setReviewPageMsg:
		m.reviewModel = newReviewModel(flashcard.NewReview(msg.Deck, m.clock), m.repository)
		m.page = Review
		return m, nil

	case setErrorPageMsg:
		m.Error = msg.Error
		m.page = Error
		return m, tea.Quit

	case setQuitPageMsg:
		m.page = Quit
		return m, tea.Quit
	}

	m, cmd = updateChildren(msg, m)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func updateChildren(msg tea.Msg, m Model) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.page {
	case Decks:
		m.deckModel, cmd = m.deckModel.Update(msg)
		return m, cmd

	case Cards:
		m.cardModel, cmd = m.cardModel.Update(m.width, msg)
		return m, cmd

	case Review:
		m.reviewModel, cmd = m.reviewModel.Update(msg)
		return m, cmd
	}

	return m, nil
}

// VIEW

func (m Model) View() string {
	switch m.page {
	case Loading:
		return m.templates.Render("loading", m)

	case Decks:
		type sizeable struct {
			Width int
			deckModel
		}
		return m.templates.Render(m.deckModel.Template(), sizeable{m.width, m.deckModel})

	case Cards:
		type sizeable struct {
			Width int
			cardModel
		}
		return m.templates.Render(m.cardModel.Template(), sizeable{m.width, m.cardModel})

	case Review:
		type sizeable struct {
			Width int
			reviewModel
		}
		return m.templates.Render(m.reviewModel.Template(), sizeable{m.width, m.reviewModel})

	case Error:
		return m.templates.Render("error", m)

	case Quit:
		return m.templates.Render("exit", m)
	}

	panic(fmt.Sprintf("missing state %d in main view", m.page))
}
