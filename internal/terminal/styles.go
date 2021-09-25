package terminal

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	DarkGreen   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "#04B575", Light: "#04B575"})
	Fuchsia     = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "#EE6FF8", Light: "#EE6FF8"})
	DarkFuchsia = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "#AD58B4", Light: "#F793FF"})
	Gray        = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "#3C3C3C", Light: "#DDDADA"})
	LightGray   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "#777777", Light: "#A49FA5"})
	Green       = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "#ECFD65", Light: "#04B575"})
	Indigo      = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "#7571F9", Light: "#5A56E0"})
	Red         = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "#ED567A", Light: "#FF4672"})
	DarkRed     = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "#C74665", Light: "#FF6F91"})
	White       = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "#DDDDDD", Light: "#1A1A1A"})

	helpStyle = lipgloss.NewStyle().Padding(1, 0, 0, 2)

	fieldStyle = lipgloss.NewStyle().Padding(1, 0)

	formStyle = lipgloss.NewStyle().Padding(1, 0)

	selectedTitleStyle = list.NewDefaultItemStyles().SelectedTitle
	selectedDescStyle  = list.NewDefaultItemStyles().SelectedDesc

	deletedDesc = deletedTitle.Copy().
			Foreground(DarkRed.GetForeground())

	deletedTitle = Red.Copy().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			Padding(0, 0, 0, 1).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"})

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Dark: "#FFFDF5", Light: "#FFFDF5"}).
			Background(lipgloss.Color("#5A56E0")).
			Padding(0, 1).
			Bold(true)
)

const HorizontalPadding = 10
