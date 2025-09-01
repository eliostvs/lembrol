package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func newErrorKeyMap() errorKeyMap {
	return errorKeyMap{
		quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
	}
}

type errorKeyMap struct {
	quit key.Binding
}

func (k errorKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.quit}
}

func (k errorKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.quit}}
}

func newErrorPage(s Shared, err error) tea.Model {
	return errorPage{
		Shared: s,
		err:    err,
		keyMap: newErrorKeyMap(),
	}
}

type errorPage struct {
	Shared
	err    error
	keyMap errorKeyMap
}

func (m errorPage) Init() tea.Cmd {
	m.Log("error: init")
	return nil
}

func (m errorPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.Log(fmt.Sprintf("error: %T", msg))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.quit):
			return m, quit
		}
	}

	return m, nil
}

func (m errorPage) View() string {
	m.Log("error: view")

	return errorView(m.Shared, m.keyMap, m.err.Error())
}

func errorView(m Shared, keyMap help.KeyMap, err string) string {
	header := m.styles.Title.
		Margin(1, 2).
		Render("Error")

	footer := lipgloss.
		NewStyle().
		Width(m.width).
		Margin(1, 2).
		Render(renderHelp(keyMap, m.width, false))

	content := m.styles.DeletedStatus.
		Width(m.width).
		Padding(0, 2).
		Height(m.height - lipgloss.Height(header) - lipgloss.Height(footer)).
		Render(err)

	return lipgloss.JoinVertical(lipgloss.Top, header, content, footer)
}
