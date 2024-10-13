package terminal

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type errorModel struct {
	Shared
	err error
}

func (m errorModel) Init() tea.Cmd {
	return nil
}

func (m errorModel) Update(tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m errorModel) View() string {
	return errorView(m.Shared, m.err.Error())
}

func errorView(m Shared, err string) string {
	header := m.styles.Title.
		Margin(1, 2).
		Render("Error")

	content := m.styles.DeletedStatus.
		Width(m.width).
		Margin(0, 2).
		Height(m.height - lipgloss.Height(header)).
		Render(err)

	return lipgloss.JoinVertical(lipgloss.Top, header, content)
}

func newErrorModel(s Shared, err error) errorModel {
	return errorModel{
		Shared: s,
		err:    err,
	}
}
