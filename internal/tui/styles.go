package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/glamour"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
)

func init() {
	var margin uint

	for _, style := range styles.DefaultStyles {
		style.Document.Margin = &margin
		style.CodeBlock.Margin = &margin
	}
}

const (
	encodedEnter = '¬'
)

var (
	fuchsia     = lipgloss.AdaptiveColor{Dark: "#EE6FF8", Light: "#EE6FF8"}
	darkFuchsia = lipgloss.AdaptiveColor{Dark: "#AD58B4", Light: "#F793FF"}
	red         = lipgloss.AdaptiveColor{Dark: "#ED567A", Light: "#FF4672"}
	darkRed     = lipgloss.AdaptiveColor{Dark: "#C74665", Light: "#FF6F91"}
	white       = lipgloss.AdaptiveColor{Dark: "#DDDDDD", Light: "#1A1A1A"}
	fieldStyle  = lipgloss.NewStyle().Foreground(white).Padding(1, 0, 0)
)

type Styles struct {
	List,
	Markdown,
	Title,
	SubTitle,
	Text,
	SelectedTitle,
	SelectedDesc,
	DeletedTitle,
	DeletedDesc,
	DeletedStatus,
	DimmedTitle lipgloss.Style
}

func NewStyles(lg *lipgloss.Renderer) *Styles {
	s := Styles{}
	s.List = lg.NewStyle().
		Padding(1, 0)
	s.Markdown = lg.NewStyle().
		Margin(0, 2)
	s.Title = lg.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Dark: "#FFFDF5", Light: "#FFFDF5"}).
		Background(lipgloss.Color("#5A56E0")).
		Padding(0, 1)
	s.SubTitle = lg.NewStyle().
		Foreground(fuchsia)
	s.Text = lg.NewStyle().Foreground(white)
	s.SelectedTitle = list.NewDefaultItemStyles().SelectedTitle
	s.SelectedDesc = list.NewDefaultItemStyles().SelectedDesc
	s.DeletedTitle = lg.NewStyle().
		Foreground(red).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}).
		Padding(0, 0, 0, 1)
	s.DeletedDesc = lg.NewStyle().
		Foreground(red).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}).
		Padding(0, 0, 0, 1).
		Foreground(darkRed)
	s.DeletedStatus = lg.NewStyle().
		Foreground(red)
	s.DimmedTitle = list.NewDefaultItemStyles().
		DimmedTitle.
		Padding(0)
	return &s
}

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

func renderHelp(keyMap help.KeyMap, width int, showAll bool) string {
	model := help.New()
	model.ShowAll = false
	model.Width = width

	return model.View(keyMap)
}
