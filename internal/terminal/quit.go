package terminal

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type quitModel struct {
	repo Repository
}

func (m quitModel) Init() tea.Cmd {
	return func() tea.Msg {
		if m.repo == nil {
			return tea.Quit()
		}

		return tea.Quit()
	}
}

func (m quitModel) Update(tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m quitModel) View() string {
	return midPaddingStyle.Render(fmt.Sprintf("Thanks for using %s!", appName))
}
