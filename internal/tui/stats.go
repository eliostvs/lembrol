package tui

import (
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

func newSparklineItem(score flashcard.ReviewScore, timestamp time.Time) sparklineItem {
	return sparklineItem{
		level:     string(levels[score-1]),
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
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case innerWindowSizeMsg:
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
		sparkline = append(sparkline, newSparklineItem(stat.Score, stat.Timestamp))
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
		Margin(1, 4).
		Render("Stats")

	subTitle := m.styles.SubTitle.
		Padding(0).
		Margin(0, 4, 1, 4).
		Render(m.card.Question)

	v := help.New()
	v.ShowAll = false
	v.Width = m.width
	footer := lipgloss.
		NewStyle().
		Width(m.width).
		Margin(1, 4).
		Render(v.View(m.keyMap))

	content := m.styles.Text.
		Width(m.width).
		Margin(0, 4).
		Height(m.height - lipgloss.Height(header) - lipgloss.Height(subTitle) - lipgloss.Height(footer)).
		Render("No Stats")

	return lipgloss.JoinVertical(lipgloss.Top, header, subTitle, content, footer)
}

func cardStatsView(m statsModel) string {
	sections := 5
	width := min(m.width/sections, 15)
	firstSession := m.sparkline[0].timestamp
	lastSession := m.sparkline[len(m.sparkline)-1].timestamp

	var content strings.Builder
	content.WriteString(m.styles.Title.Render("Stats"))
	content.WriteString(m.styles.SubTitle.Render(m.card.Question))
	content.WriteString("\n")
	content.WriteString(m.styles.Text.Copy().Align(lipgloss.Left).Render(firstSession.Format("02/01/2006")))
	content.WriteString(m.styles.Text.Copy().Width(width * (sections - 1)).Align(lipgloss.Right).Render(lastSession.Format("02/01/2006")))
	content.WriteString("\n")

	headerStyle := lipgloss.NewStyle().Foreground(darkFuchsia).Width(width).Align(lipgloss.Left)
	for _, header := range []string{"TOTAL", "HARD", "NORMAL", "EASY", "VERY EASY"} {
		content.WriteString(headerStyle.Render(header))
	}
	content.WriteString("\n")

	totalStyle := m.styles.Text.Copy().Width(width).Align(lipgloss.Left)
	totals := []flashcard.ReviewScore{
		flashcard.ReviewScoreAgain,
		flashcard.ReviewScoreHard,
		flashcard.ReviewScoreNormal,
		flashcard.ReviewScoreEasy,
		flashcard.ReviewScoreSuperEasy,
	}
	for _, total := range totals {
		content.WriteString(totalStyle.Render(strconv.Itoa(m.totals[total])))
	}
	content.WriteString("\n\n")

	for _, item := range m.sparkline {
		content.WriteString(item.level)
	}
	content.WriteString("\n")

	content.WriteString(renderHelp(m.keyMap, m.width, m.height-lipgloss.Height(content.String()), false))
	return m.styles.Margin.Render(content.String())
}
