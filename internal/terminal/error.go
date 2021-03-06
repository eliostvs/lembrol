package terminal

import (
	tea "github.com/charmbracelet/bubbletea"
)

type errorModel struct {
	err error
}

func (m errorModel) Init() tea.Cmd {
	return nil
}

func (m errorModel) Update(tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m errorModel) View() string {
	return errorView(m.err.Error())
}

func errorView(err string) string {
	content := titleStyle.Render("Error")
	content += Red.Render(err)
	return largePaddingStyle.Render(content)
}
