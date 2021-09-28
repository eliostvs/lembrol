package terminal

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
)

const (
	inputPrompt = "> "

	encodedEnter = "¬"
)

var (
	reviewScreenStyle = lipgloss.NewStyle().Padding(1, 0)

	midPaddingStyle = lipgloss.NewStyle().Padding(1, 2)

	largePaddingStyle = lipgloss.NewStyle().Padding(1, 4)

	Fuchsia     = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "#EE6FF8", Light: "#EE6FF8"})
	DarkFuchsia = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "#AD58B4", Light: "#F793FF"})
	Red         = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "#ED567A", Light: "#FF4672"})
	DarkRed     = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "#C74665", Light: "#FF6F91"})
	White       = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "#DDDDDD", Light: "#1A1A1A"})

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Dark: "#FFFDF5", Light: "#FFFDF5"}).
			Background(lipgloss.Color("#5A56E0")).
			Padding(0, 1).
			Bold(true)

	titleReviewStyle = titleStyle.Copy().Margin(0, 4)

	helpStyle = lipgloss.NewStyle().Margin(2, 0, 0)

	status = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Dark: "#3C3C3C", Light: "#DDDADA"}).
		Margin(1, 4)

	deckName = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Dark: "#7571F9", Light: "#5A56E0"}).
			Margin(2, 4, 0)

	markdownStyle = lipgloss.NewStyle().Padding(1, 1)

	helpReviewStyle = lipgloss.NewStyle().Margin(1, 4)

	normalTextStyle = White.Copy().Margin(2, 0)

	fieldStyle = lipgloss.NewStyle().Margin(1, 0, 0)

	formStyle = lipgloss.NewStyle().Margin(1, 0, 2)

	selectedTitleStyle = list.NewDefaultItemStyles().SelectedTitle

	selectedDescStyle = list.NewDefaultItemStyles().SelectedDesc

	deletedDesc = deletedTitle.Copy().
			Foreground(DarkRed.GetForeground())

	deletedTitle = Red.Copy().
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

func RenderMarkdown(in string, width int) (string, error) {
	r, _ := glamour.NewTermRenderer(
		glamour.WithEnvironmentConfig(),
		glamour.WithWordWrap(width),
	)

	lines, err := r.Render(renderMultiline(in, ""))
	if err != nil {
		return "", err
	}

	return strings.ReplaceAll(lines, "\n", "\n "), nil
}

func renderMultiline(content, prefix string) string {
	return strings.Replace(content, encodedEnter, "\n"+prefix, -1)
}

func pluralize(val int, suffix string) string {
	if val > 1 {
		return suffix
	}
	return ""
}
