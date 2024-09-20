package terminal

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func newQuitModel(s Shared) quitModel {
	return quitModel{
		Shared: s,
	}
}

type quitModel struct {
	Shared
}

func (m quitModel) Init() tea.Cmd {
	return func() tea.Msg {
		return tea.Quit()
	}
}

func (m quitModel) Update(tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m quitModel) View() string {
	header := m.styles.Title.
		Margin(1, 2).
		Render("Bye")

	content := m.styles.Text.
		Width(m.width).
		Margin(0, 2).
		Height(m.height - lipgloss.Height(header)).
		Render(fmt.Sprintf("Thanks for using %s!", appName))

	return lipgloss.JoinVertical(lipgloss.Top, header, content)
}
