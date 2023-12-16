package terminal

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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
	return errorView(m.styles, m.err.Error())
}

func errorView(styles *Styles, err string) string {
	var content strings.Builder
	content.WriteString(styles.Title.Render("Error"))
	content.WriteString(styles.DeletedStatus.Render(err))
	return styles.Pagination.Render(content.String())
}

func newErrorModel(s Shared, err error) errorModel {
	return errorModel{
		Shared: s,
		err:    err,
	}
}
