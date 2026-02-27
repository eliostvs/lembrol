package tui

import (
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/lipgloss/v2/compat"
	"github.com/charmbracelet/glamour"

	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/glamour/styles"
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
	fuchsia     = compat.AdaptiveColor{Dark: lipgloss.Color("#EE6FF8"), Light: lipgloss.Color("#EE6FF8")}
	darkFuchsia = compat.AdaptiveColor{Dark: lipgloss.Color("#AD58B4"), Light: lipgloss.Color("#F793FF")}
	red         = compat.AdaptiveColor{Dark: lipgloss.Color("#ED567A"), Light: lipgloss.Color("#FF4672")}
	darkRed     = compat.AdaptiveColor{Dark: lipgloss.Color("#C74665"), Light: lipgloss.Color("#FF6F91")}
	white       = compat.AdaptiveColor{Dark: lipgloss.Color("#DDDDDD"), Light: lipgloss.Color("#1A1A1A")}
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

func NewStyles() *Styles {
	s := Styles{}
	s.List = lipgloss.NewStyle().
		Padding(1, 0)
	s.Markdown = lipgloss.NewStyle().
		Margin(0, 2)
	s.Title = lipgloss.NewStyle().
		Foreground(compat.AdaptiveColor{Dark: lipgloss.Color("#FFFDF5"), Light: lipgloss.Color("#FFFDF5")}).
		Background(lipgloss.Color("#5A56E0")).
		Padding(0, 1)
	s.SubTitle = lipgloss.NewStyle().
		Foreground(fuchsia)
	s.Text = lipgloss.NewStyle().Foreground(white)
	s.SelectedTitle = list.NewDefaultItemStyles(true).SelectedTitle
	s.SelectedDesc = list.NewDefaultItemStyles(true).SelectedDesc
	s.DeletedTitle = lipgloss.NewStyle().
		Foreground(red).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(compat.AdaptiveColor{Light: lipgloss.Color("#F793FF"), Dark: lipgloss.Color("#AD58B4")}).
		Padding(0, 0, 0, 1)
	s.DeletedDesc = lipgloss.NewStyle().
		Foreground(red).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(compat.AdaptiveColor{Light: lipgloss.Color("#F793FF"), Dark: lipgloss.Color("#AD58B4")}).
		Padding(0, 0, 0, 1).
		Foreground(darkRed)
	s.DeletedStatus = lipgloss.NewStyle().
		Foreground(red)
	s.DimmedTitle = list.NewDefaultItemStyles(true).
		DimmedTitle.
		Padding(0)
	return &s
}

func naturalTime(t time.Time) string {
	if time.Since(t) < time.Minute {
		return "just now"
	}
	return humanize.Time(t)
}

func RenderMarkdown(text string, width int) (string, error) {
	r, _ := glamour.NewTermRenderer(
		glamour.WithWordWrap(width),
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

func renderHelp(keyMap help.KeyMap, width int, fullHelp bool) string {
	model := help.New()
	model.SetWidth(width)
	model.ShowAll = fullHelp

	return model.View(keyMap)
}
