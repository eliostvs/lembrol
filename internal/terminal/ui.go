package terminal

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/eliostvs/remembercli/internal/flashcard"
)

const (
	initialDelay = time.Millisecond * 800
)

// MODEL

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

func newWindowSize(style lipgloss.Style, msg tea.WindowSizeMsg) windowSize {
	topGap, rightGap, bottomGap, leftGap := style.GetPadding()
	return windowSize{
		width:  msg.Width - leftGap - rightGap - 2,
		height: msg.Height - topGap - bottomGap,
	}
}

// windowSize is the size of terminal minus the edges paddings.
type windowSize struct {
	width, height int
}

// NewModel creates a new model instance given a decks location.
func NewModel(location string, opts ...ModelOption) Model {
	spin := spinner.NewModel()
	spin.Spinner = spinner.Dot

	m := Model{
		spinner:      spin,
		clock:        flashcard.NewClock(),
		initialDelay: initialDelay,
		location:     location,
	}

	for _, opt := range opts {
		opt(&m)
	}

	return m
}

type Model struct {
	error        string
	spinner      spinner.Model
	cardsModel   cardsModel
	clock        flashcard.Clock
	decksModel   decksModel
	initialDelay time.Duration
	location     string
	page         page
	repository   *flashcard.Repository
	reviewModel  reviewModel
	window       windowSize
}

// VIEW

func (m Model) View() string {
	switch m.page {
	case Loading:
		return loadingView(m)

	case Decks:
		return m.decksModel.View(m.window)

	case Cards:
		return m.cardsModel.View(m.window)

	case Review:
		return m.reviewModel.View(m.window)

	case Error:
		return errorView(m.error)

	case Quit:
		return midPaddingStyle.Render("Thanks for using Remember CLI!")
	}

	panic(midPaddingStyle.Render(fmt.Sprintf("missing state %d in main view", m.page)))
}

func loadingView(m Model) string {
	content := titleStyle.Render("Remember")
	content += normalTextStyle.Render(fmt.Sprintf("%s Loading...", m.spinner.View()))
	return largePaddingStyle.Render(content)
}

func errorView(err string) string {
	content := titleStyle.Render("Error")
	content += Red.Render(err)
	return largePaddingStyle.Render(content)
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
		m.window = newWindowSize(largePaddingStyle, msg)
		return m, nil

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case createdRepositoryMsg:
		m.repository = msg.Repository
		m.decksModel = newDecksModel(m.repository.List(), m.repository)
		m.page = Decks
		return m, m.decksModel.init()

	case setDecksPageMsg:
		m.page = Decks
		return m, m.decksModel.init()

	case setDeckPageMsg:
		m.cardsModel = newCardsModel(msg.Deck, m.clock, m.repository)
		m.page = Cards
		return m, m.cardsModel.init()

	case setReviewPageMsg:
		m.reviewModel = newReviewModel(flashcard.NewReview(msg.Deck, m.clock), m.repository)
		m.page = Review
		return m, m.reviewModel.init()

	case setErrorPageMsg:
		m.error = msg.Error
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
		m.decksModel, cmd = m.decksModel.Update(msg)
		return m, cmd

	case Cards:
		m.cardsModel, cmd = m.cardsModel.Update(m.window, msg)
		return m, cmd

	case Review:
		m.reviewModel, cmd = m.reviewModel.Update(msg)
		return m, cmd
	}

	return m, nil
}
