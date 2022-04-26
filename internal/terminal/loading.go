package terminal

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func newLoadinModel(title string) loadingModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return loadingModel{
		spinner: s,
		title:   title,
	}

}

type loadingModel struct {
	title   string
	spinner spinner.Model
}

func (loadingModel) Init() tea.Cmd {
	return nil
}

func (m loadingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m loadingModel) View() string {
	return loadingView(m.title, m.spinner)
}
