package tui

import (
	"fmt"
	"log"
	"time"

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
		m.clock.Sleep(time.Second)
		return tea.Quit()
	}
}

func (m quitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Printf("quit: %T\n", msg)

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	}

	return m, cmd
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
