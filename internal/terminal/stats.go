package terminal

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/eliostvs/lembrol/internal/flashcard"
)

// MODEL

var levels = []rune("▁▃▅█")

type statsKeys struct {
	cancel key.Binding
}

func (k statsKeys) ShortHelp() []key.Binding {
	return []key.Binding{
		k.cancel,
	}
}

func (k statsKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.cancel}}
}

type statsState int

const (
	statsLoading statsState = iota
	statsLoaded
)

func newStatsModel(msg setStatsPageMsg, repo *flashcard.Repository, width, height int) statsModel {
	spin := spinner.New()
	spin.Spinner = spinner.Dot

	helpModel := help.New()
	helpModel.Width = width

	return statsModel{
		card:       msg.card,
		cardIndex:  msg.cardIndex,
		deck:       msg.deck,
		repository: repo,
		spinner:    spin,
		state:      statsLoading,
		keys: statsKeys{
			key.NewBinding(
				key.WithKeys("q", tea.KeyEsc.String()),
				key.WithHelp("q", "quit"),
			),
		},
		width:  width,
		height: height,
		totals: make(map[flashcard.ReviewScore]int),
		help:   helpModel,
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
	card       flashcard.Card
	cardIndex  int
	deck       flashcard.Deck
	keys       statsKeys
	repository *flashcard.Repository
	spinner    spinner.Model
	state      statsState
	totals     map[flashcard.ReviewScore]int
	sparkline  []sparklineItem
	help       help.Model
	width      int
	height     int
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
				return loadStats(m.repository, m.deck, m.card)
			},
		),
		spinner.Tick,
	)
}

func (m statsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case innerWindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.help.Width = msg.Width
		return m, nil

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case statsLoadedMsg:
		m.state = statsLoaded
		m.sparkline = createSparkline(msg.stats)
		m.totals = calculateTotals(msg.stats)
		return m, cmd

	case tea.KeyMsg:
		if key.Matches(msg, m.keys.cancel) {
			return m, showCards(m.deck, m.cardIndex)
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

func loadStats(repo *flashcard.Repository, deck flashcard.Deck, card flashcard.Card) tea.Msg {
	stats, err := repo.Stats.Find(deck, card)
	if err != nil {
		return fail(err)
	}

	return statsLoadedMsg{stats: stats}
}

// VIEW

func (m statsModel) View() string {
	switch m.state {
	case statsLoading:
		return loadingView("Stats", m.spinner)

	case statsLoaded:
		if len(m.sparkline) > 0 {
			return cardStatsView(m)
		}
		return notStatsView(m)

	default:
		return ""
	}
}

func loadingView(title string, spin spinner.Model) string {
	content := titleStyle.Render(title)
	content += normalTextStyle.Render(fmt.Sprintf("%s Loading...", spin.View()))
	return largePaddingStyle.Render(content)
}

func notStatsView(m statsModel) string {
	content := titleStyle.Render("Stats")
	content += Fuchsia.Copy().Margin(2, 0, 1).Render(m.card.Question)
	content += "\n"
	content += White.Render("No stats")
	content += "\n"
	content += helpStyle.Render(m.help.View(m.keys))
	return largePaddingStyle.Render(content)
}

func cardStatsView(m statsModel) string {
	sections := 5
	width := min(m.width/sections, 15)
	firstSession := m.sparkline[0].timestamp
	lastSession := m.sparkline[len(m.sparkline)-1].timestamp

	content := titleStyle.Render("Stats")
	content += Fuchsia.Copy().Margin(2, 0, 1).Render(m.card.Question)
	content += "\n\n"
	content += White.Copy().Align(lipgloss.Left).Render(firstSession.Format("02/01/2006"))
	content += White.Copy().Width(width * (sections - 1)).Align(lipgloss.Right).Render(lastSession.Format("02/01/2006"))
	content += "\n\n"

	var headerStyle = DarkFuchsia.Copy().Width(width).Align(lipgloss.Left)
	for _, header := range []string{"TOTAL", "HARD", "NORMAL", "EASY", "VERY EASY"} {
		content += headerStyle.Render(header)
	}
	content += "\n"

	var totalStyle = White.Copy().Width(width).Align(lipgloss.Left)
	totals := []flashcard.ReviewScore{
		flashcard.ReviewScoreAgain,
		flashcard.ReviewScoreHard,
		flashcard.ReviewScoreNormal,
		flashcard.ReviewScoreEasy,
		flashcard.ReviewScoreSuperEasy,
	}
	for _, total := range totals {
		content += totalStyle.Render(strconv.Itoa(m.totals[total]))
	}
	content += "\n\n"

	for _, item := range m.sparkline {
		content += item.level
	}
	content += "\n"

	content += helpStyle.Render(m.help.View(m.keys))
	return largePaddingStyle.Render(content)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
