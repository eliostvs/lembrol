package terminal

import (
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
		m.width, m.height = calcInnerWindowSize(
			largePaddingStyle, tea.WindowSizeMsg{
				Width:  width,
				Height: height,
			},
		)
	}
}

func calcInnerWindowSize(style lipgloss.Style, msg tea.WindowSizeMsg) (int, int) {
	topGap, rightGap, bottomGap, leftGap := style.GetPadding()
	return msg.Width - leftGap - rightGap - 2, msg.Height - topGap - bottomGap
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
func NewModel(path string, opts ...ModelOption) Model {
	m := Model{
		page: newLoadingPage(appName, "Loading...", appCommon{}),
		repositoryFactory: func(c clock.Clock) (Repository, error) {
			return flashcard.NewRepository(path, c)
		},
		appCommon: appCommon{
			clock: clock.New(),
		},
	}

	for _, opt := range opts {
		opt(&m)
	}

	return m
}

type appCommon struct {
	repository Repository
	clock      clock.Clock
	width      int
	height     int
}

type Model struct {
	repositoryFactory func(clock.Clock) (Repository, error)
	page              tea.Model
	appCommon
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
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = calcInnerWindowSize(largePaddingStyle, msg)
		return m, changeInnerWindowSize(m.width, m.height)

	case createdRepositoryMsg:
		m.repository = msg.repository
		m.page = newDeckPage(0, m.appCommon)
		return m, m.page.Init()

	case setDecksPageMsg:
		m.page = newDeckPage(0, m.appCommon)
		return m, m.page.Init()

	case setCardsPageMsg:
		m.page = newCardPage(msg.deck, m.appCommon)
		return m, m.page.Init()

	case setStatsPageMsg:
		m.page = newStatsModel(msg, m.appCommon)
		return m, m.page.Init()

	case setReviewPageMsg:
		m.page = newReviewPage(flashcard.NewReview(msg.Deck, m.clock), m.appCommon)
		return m, m.page.Init()

	case setErrorPageMsg:
		m.page = errorModel{msg.err}
		return m, tea.Quit

	case setQuitPageMsg:
		m.page = quitModel{m.repository}
		return m, m.page.Init()
	}

	m.page, cmd = m.page.Update(msg)

	return m, cmd
}

// VIEW

func (m Model) View() string {
	return m.page.View()
}
