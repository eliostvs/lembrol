package terminal

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/eliostvs/lembrol/internal/clock"
	"github.com/eliostvs/lembrol/internal/flashcard"
)

const (
	initialDelay = time.Millisecond * 600
	projectName  = "Lembrol"
)

// MODEL

type ModelOption func(*Model)

func WithWindowSize(width, height int) ModelOption {
	return func(m *Model) {
		m.width, m.height = calcInnerWindowSize(
			largePaddingStyle, tea.WindowSizeMsg{
				Width:  width,
				Height: height,
			},
		)
	}
}

func WithInitialDelay(delay time.Duration) ModelOption {
	return func(m *Model) {
		m.initialDelay = delay
	}
}

func WithClock(clock clock.Clock) ModelOption {
	return func(m *Model) {
		m.clock = clock
	}
}

func calcInnerWindowSize(style lipgloss.Style, msg tea.WindowSizeMsg) (int, int) {
	topGap, rightGap, bottomGap, leftGap := style.GetPadding()
	return msg.Width - leftGap - rightGap - 2, msg.Height - topGap - bottomGap
}

func createRepository(location string, clock clock.Clock) tea.Msg {
	repo, err := flashcard.NewRepository(location, clock)
	if err != nil {
		return fail(err)
	}
	return createdRepositoryMsg{repo}
}

type createdRepositoryMsg struct {
	*flashcard.Repository
}

// NewModel creates a new model instance given a decks location.
func NewModel(location string, opts ...ModelOption) Model {
	m := Model{
		clock:        clock.New(),
		initialDelay: initialDelay,
		location:     location,
		page:         newLoadinModel(projectName),
	}

	for _, opt := range opts {
		opt(&m)
	}

	return m
}

type Model struct {
	clock        clock.Clock
	initialDelay time.Duration
	location     string
	repository   *flashcard.Repository
	page         tea.Model
	width        int
	height       int
}

// MESSAGES

func changeInnerWindowSize(width, height int) tea.Cmd {
	return func() tea.Msg {
		return innerWindowSizeMsg{Width: width, Height: height}
	}
}

type innerWindowSizeMsg struct {
	Width, Height int
}

func fail(err error) tea.Msg {
	return setErrorPageMsg{err}
}

type setErrorPageMsg struct {
	err error
}

func showDecks() tea.Msg {
	return setDecksPageMsg{}
}

type setDecksPageMsg struct{}

func showCards(deck flashcard.Deck, cardIndex int) tea.Cmd {
	return func() tea.Msg {
		return setCardsPageMsg{deck: deck, cardIndex: cardIndex}
	}
}

type setCardsPageMsg struct {
	deck      flashcard.Deck
	cardIndex int
}

func showStats(deck flashcard.Deck, card flashcard.Card, cardIndex int) tea.Cmd {
	return func() tea.Msg {
		return setStatsPageMsg{deck: deck, card: card, cardIndex: cardIndex}
	}
}

type setStatsPageMsg struct {
	deck      flashcard.Deck
	card      flashcard.Card
	cardIndex int
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
	return tea.Batch(
		tea.Tick(
			m.initialDelay,
			func(time.Time) tea.Msg {
				return createRepository(m.location, m.clock)
			},
		),
		spinner.Tick,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = calcInnerWindowSize(largePaddingStyle, msg)
		return m, changeInnerWindowSize(m.width, m.height)

	case createdRepositoryMsg:
		m.repository = msg.Repository
		m.page = newDecksModel(m.repository, m.width, m.height)
		return m, m.page.Init()

	case setDecksPageMsg:
		m.page = newDecksModel(m.repository, m.width, m.height)
		return m, m.page.Init()

	case setCardsPageMsg:
		m.page = newCardsModel(msg, m.clock, m.repository, m.width, m.height)
		return m, m.page.Init()

	case setStatsPageMsg:
		m.page = newStatsModel(msg, m.repository, m.width, m.height)
		return m, m.page.Init()

	case setReviewPageMsg:
		m.page = newReviewModel(flashcard.NewReview(msg.Deck, m.clock), m.repository, m.width, m.height)
		return m, m.page.Init()

	case setErrorPageMsg:
		m.page = errorModel{msg.err}
		return m, tea.Quit

	case setQuitPageMsg:
		m.page = quitModel{}
		return m, tea.Quit
	}

	m.page, cmd = m.page.Update(msg)

	return m, cmd
}

// VIEW

func (m Model) View() string {
	return m.page.View()
}
