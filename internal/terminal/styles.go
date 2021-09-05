package terminal

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
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
)

const HorizontalPadding = 10

func Width(msg tea.WindowSizeMsg) int {
	return msg.Width - HorizontalPadding
}
