package terminal

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
)

func init() {
	var margin uint

	for _, style := range glamour.DefaultStyles {
		style.Document.Margin = &margin
		style.CodeBlock.Margin = &margin
	}
}

const (
	encodedEnter = '¬'
)

var (
	Fuchsia     = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "#EE6FF8", Light: "#EE6FF8"})
	DarkFuchsia = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "#AD58B4", Light: "#F793FF"})
	Red         = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "#ED567A", Light: "#FF4672"})
	DarkRed     = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "#C74665", Light: "#FF6F91"})
	White       = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "#DDDDDD", Light: "#1A1A1A"})

	midPaddingStyle   = lipgloss.NewStyle().Padding(1, 2)
	largePaddingStyle = lipgloss.NewStyle().Padding(1, 4)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Dark: "#FFFDF5", Light: "#FFFDF5"}).
			Background(lipgloss.Color("#5A56E0")).
			Padding(0, 1)

	subTitleStyle = Fuchsia.
			Copy().
			Padding(2, 0, 1)

	paginationStyle = DarkFuchsia.
			Copy().
			Margin(1, 0)

	normalTextStyle = list.NewDefaultItemStyles().
			NormalTitle.
			Copy().
			Padding(2, 0)

	selectedTitleStyle = list.NewDefaultItemStyles().SelectedTitle

	selectedDescStyle = list.NewDefaultItemStyles().SelectedDesc

	deletedDescStyle = deletedTitleStyle.Copy().
				Foreground(DarkRed.GetForeground())

	deletedTitleStyle = Red.Copy().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				Padding(0, 0, 0, 1).
				BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"})
)

func naturalTime(t time.Time) string {
	if time.Now().Sub(t) < time.Minute {
		return "just now"
	}
	return humanize.Time(t)
}

func RenderMarkdown(text string, width int) (string, error) {
	r, _ := glamour.NewTermRenderer(
		glamour.WithWordWrap(width),
		glamour.WithEnvironmentConfig(),
	)

	lines, err := r.Render(breakLines(text))
	if err != nil {
		return "", err
	}

	return lines, nil
}

func breakLines(content string) string {
	return strings.ReplaceAll(content, string(encodedEnter), "\n")
}

func pluralize(val int, suffix string) string {
	if val > 1 {
		return suffix
	}
	return ""
}

func renderHelp(keyMap help.KeyMap, width, height int, showAll bool) string {
	m := help.New()
	m.ShowAll = showAll
	if width > 0 {
		m.Width = width
	}

	helpText := m.View(keyMap)
	minHeight := lipgloss.Height(helpText) - 1
	var prefix string

	if height > minHeight {
		prefix += strings.Repeat("\n", height-minHeight)
	}

	return prefix + helpText
}
