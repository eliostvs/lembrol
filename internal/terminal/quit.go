package terminal

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type quitModel struct {
}

func (m quitModel) Init() tea.Cmd {
	return nil
}

func (m quitModel) Update(tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m quitModel) View() string {
	return midPaddingStyle.Render(fmt.Sprintf("Thanks for using %s!", projectName))
}
