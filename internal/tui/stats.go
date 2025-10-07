package tui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eliostvs/lembrol/internal/flashcard"
)

// Model

var levels = []rune("▁▃▅█")

type statsKeyMap struct {
	cancel key.Binding
}

func (k statsKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.cancel,
	}
}

func (k statsKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.cancel}}
}

type statsState int

const (
	statsLoading statsState = iota
	statsLoaded
)

func newStatsModel(shared Shared, msg setStatsPageMsg) statsModel {
	return statsModel{
		Shared:    shared,
		card:      msg.card,
		cardIndex: msg.cardIndex,
		deck:      msg.deck,
		loading:   newLoadingPage(shared, "Stats", "Loading..."),
		state:     statsLoading,
		keyMap: statsKeyMap{
			key.NewBinding(
				key.WithKeys("q", tea.KeyEsc.String()),
				key.WithHelp("q", "quit"),
			),
		},
		totals: make(map[flashcard.ReviewScore]int),
	}
}

func newSparklineItem(s flashcard.ReviewScore, timestamp time.Time) sparklineItem {
	return sparklineItem{
		level:     string(rune(levels[s-1])),
		timestamp: timestamp,
	}
}

type sparklineItem struct {
	timestamp time.Time
	level     string
}

type statsModel struct {
	Shared
	card      flashcard.Card
	cardIndex int
	deck      flashcard.Deck
	keyMap    statsKeyMap
	loading   tea.Model
	state     statsState
	totals    map[flashcard.ReviewScore]int
	sparkline []sparklineItem
}

// MESSAGES

type (
	statsLoadedMsg struct {
		stats []flashcard.Stats
	}
)

// UPDATE

func (m statsModel) Init() tea.Cmd {
	m.Log("stats: init")

	return tea.Batch(
		tea.Tick(
			time.Millisecond*500, func(time.Time) tea.Msg {
				return loadStats(m.card)
			},
		),
		m.loading.Init(),
	)
}

func (m statsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.Log(fmt.Sprintf("stats: %T", msg))

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case spinner.TickMsg:
		if m.state != statsLoading {
			return m, nil
		}
		m.loading, cmd = m.loading.Update(msg)
		return m, cmd

	case statsLoadedMsg:
		m.state = statsLoaded
		m.sparkline = createSparkline(msg.stats)
		m.totals = calculateTotals(msg.stats)
		return m, cmd

	case tea.KeyMsg:
		if key.Matches(msg, m.keyMap.cancel) {
			return m, showCards(m.cardIndex, m.deck)
		}
	}

	return m, nil
}

func calculateTotals(stats []flashcard.Stats) map[flashcard.ReviewScore]int {
	totals := make(map[flashcard.ReviewScore]int, 5)
	for _, stat := range stats {
		totals[flashcard.ReviewScore(0)]++ // total
		totals[stat.Score]++
	}
	return totals
}

func createSparkline(stats []flashcard.Stats) []sparklineItem {
	sparkline := make([]sparklineItem, 0, len(stats))

	for _, stat := range stats {
		sparkline = append(sparkline, newSparklineItem(stat.Score, stat.LastReview))
	}

	sort.Slice(
		sparkline, func(i, j int) bool {
			return sparkline[i].timestamp.Before(sparkline[j].timestamp)
		},
	)

	return sparkline
}

func loadStats(card flashcard.Card) tea.Msg {
	return statsLoadedMsg{stats: card.Stats}
}

// VIEW

func (m statsModel) View() string {
	m.Log("stats: view")

	switch m.state {
	case statsLoading:
		return m.loading.View()

	case statsLoaded:
		if len(m.sparkline) > 0 {
			return cardStatsView(m)
		}
		return notStatsView(m)

	default:
		return ""
	}
}

func notStatsView(m statsModel) string {
	header := m.styles.Title.
		Margin(1, 2).
		Render("Stats")

	subTitle := m.styles.SubTitle.
		Margin(0, 2, 1).
		Render(m.card.Question)

	v := help.New()
	v.ShowAll = false
	v.Width = m.width
	footer := lipgloss.
		NewStyle().
		Width(m.width).
		Margin(1, 2).
		Render(v.View(m.keyMap))

	content := m.styles.Text.
		Width(m.width).
		Height(m.height-lipgloss.Height(header)-lipgloss.Height(subTitle)-lipgloss.Height(footer)).
		Margin(0, 2).
		Render("No stats")

	return lipgloss.JoinVertical(lipgloss.Top, header, subTitle, content, footer)
}

func cardStatsView(m statsModel) string {
	sections := 5
	width := min(m.width/sections, 12)

	title := m.styles.Title.
		Margin(1, 2).
		Render("Stats")

	question := m.styles.SubTitle.
		Margin(0, 2, 0).
		Render(m.card.Question)

	margins := 4
	firstSession := m.styles.Text.
		Width(width*(sections-1)+margins*(sections-2)).
		Margin(1, 2).
		Align(lipgloss.Left).
		Render(m.sparkline[0].timestamp.Format("02/01/2006"))
	lastSession := m.styles.Text.
		Width(width).
		Margin(1, 2).
		Align(lipgloss.Left).
		Render(m.sparkline[len(m.sparkline)-1].timestamp.Format("02/01/2006"))
	dates := lipgloss.JoinHorizontal(lipgloss.Left, firstSession, lastSession)

	headerStyle := lipgloss.NewStyle().
		Width(width).
		Margin(0, 2).
		Foreground(darkFuchsia).
		Align(lipgloss.Left)
	scoreLabels := make([]string, sections)
	for i, label := range []string{"TOTAL", "AGAIN", "HARD", "NORMAL", "EASY"} {
		scoreLabels[i] = headerStyle.Render(label)
	}
	totalLabels := lipgloss.JoinHorizontal(lipgloss.Left, scoreLabels...)

	totalStyle := m.styles.Text.
		Width(width).
		Margin(0, 2).
		Align(lipgloss.Left)
	scoreTotals := make([]string, sections)
	for i := range sections {
		scoreTotals[i] = totalStyle.Render(strconv.Itoa(m.totals[flashcard.ReviewScore(i)]))
	}
	totals := lipgloss.JoinHorizontal(lipgloss.Left, scoreTotals...)

	actions := lipgloss.
		NewStyle().
		Width(m.width).
		Margin(1, 2, 0).
		Render(renderHelp(m.keyMap, m.width, false))

	var content strings.Builder
	for _, item := range m.sparkline {
		content.WriteString(item.level)
	}
	sparkline := m.styles.Text.
		Width(m.width).
		Margin(1, 2).
		Align(lipgloss.Left).
		Height(m.height - lipgloss.Height(title) - lipgloss.Height(question) - lipgloss.Height(dates) - lipgloss.Height(totalLabels) - lipgloss.Height(dates) - lipgloss.Height(totals) - lipgloss.Height(actions)).
		Render(content.String())

	return lipgloss.JoinVertical(lipgloss.Top, title, question, dates, totalLabels, totals, sparkline, actions)
}
