package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func showLoading(title, description string) tea.Cmd {
	return func() tea.Msg {
		return showLoadingMsg{title: title, description: description}
	}
}

type showLoadingMsg struct {
	title       string
	description string
}

type loadingKeyMap struct {
	forceQuit key.Binding
}

func (k loadingKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.forceQuit}
}

func (k loadingKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.forceQuit}}
}

func newLoadingPage(shared Shared, title, description string) loadingPage {
	return loadingPage{
		spinner:     spinner.New(spinner.WithSpinner(spinner.Dot)),
		title:       title,
		description: description,
		Shared:      shared,
		keyMap: loadingKeyMap{
			forceQuit: key.NewBinding(
				key.WithKeys("ctrl+c"),
				key.WithHelp("ctrl+c", "quit"),
			),
		},
	}
}

type loadingPage struct {
	Shared
	title       string
	description string
	spinner     spinner.Model
	keyMap      loadingKeyMap
}

func (m loadingPage) Init() tea.Cmd {
	m.Log("loading: init")

	return m.spinner.Tick
}

func (m loadingPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.Log(fmt.Sprintf("loading: %T", msg))

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.forceQuit):
			return m, quit
		}
	}

	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m loadingPage) View() string {
	m.Log("loading: view")

	header := m.styles.Title.
		Margin(1, 2).
		Render(m.title)

	footer := lipgloss.
		NewStyle().
		Width(m.width).
		Margin(1, 2).
		Render(renderHelp(m.keyMap, m.width, false))

	content := m.styles.Text.
		Width(m.width).
		Margin(0, 2).
		Height(m.height - lipgloss.Height(header) - lipgloss.Height(footer)).
		Render(fmt.Sprintf("%s %s", m.spinner.View(), m.description))

	return lipgloss.JoinVertical(lipgloss.Top, header, content, footer)
}
