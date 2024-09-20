package terminal

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
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
	return m.styles.ListMargin.Render(fmt.Sprintf("Thanks for using %s!", appName))
}
