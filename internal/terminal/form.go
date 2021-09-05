package terminal

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func newField(name string, model textinput.Model) field {
	return field{name: name, model: model}
}

type field struct {
	name  string
	model textinput.Model
}

func (f field) Focus() (field, tea.Cmd) {
	f.model.PromptStyle = Fuchsia
	f.model.TextStyle = Fuchsia
	return f, f.model.Focus()
}

func (f field) Blur() field {
	f.model.PromptStyle = DarkFuchsia
	f.model.TextStyle = DarkFuchsia
	f.model.Blur()
	return f
}

func (f field) Update(msg tea.Msg) (field, tea.Cmd) {
	var cmd tea.Cmd
	f.model, cmd = f.model.Update(msg)
	return f, cmd
}

func (f field) Match(name string) bool {
	return f.name == name
}

func (f field) IsValid() bool {
	return 0 < len(f.model.Value())
}

func (f field) View() string {
	return f.model.View()
}

func (f field) Value() string {
	return f.model.Value()
}

// Form Type

func submit(f form) tea.Cmd {
	return func() tea.Msg {
		return submittedFormMsg{f}
	}
}

type submittedFormMsg struct {
	Form form
}

func newForm(f field, fields ...field) form {
	return form{
		cursor: newCursor(len(fields)),
		fields: append([]field{f}, fields...),
	}
}

type form struct {
	cursor cursor
	fields []field
}

func (f form) Focus(index int) (form, tea.Cmd) {
	var cmd tea.Cmd

	for i := range f.fields {
		if i == index {
			f.fields[i], cmd = f.fields[i].Focus()
		} else {
			f.fields[i] = f.fields[i].Blur()
		}
	}

	return f, cmd
}

func (f form) Width(width int) form {
	for i := range f.fields {
		f.fields[i].model.Width = width
	}

	return f
}

func (f form) Error(name string) bool {
	for _, field := range f.fields {
		if field.Match(name) {
			return field.IsValid()
		}
	}
	return false
}

func (f form) isValid() bool {
	for _, field := range f.fields {
		if !field.IsValid() {
			return false
		}
	}
	return true
}

func (f form) Value(name string) string {
	for _, field := range f.fields {
		if field.Match(name) {
			return field.Value()
		}
	}
	return ""
}

func (f form) View(name string) string {
	for _, field := range f.fields {
		if field.Match(name) {
			return field.View()
		}
	}
	return ""
}

func (f form) Prev() (form, tea.Cmd) {
	f.cursor.Up()
	return f.Focus(f.cursor.Value())
}

func (f form) Next() (form, tea.Cmd) {
	f.cursor.Down()
	return f.Focus(f.cursor.Value())
}

func (f form) Update(msg tea.Msg) (form, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "shift+tab", tea.KeyUp.String():
			return f.Prev()

		case "tab", tea.KeyDown.String():
			return f.Next()

		case tea.KeyEnter.String():
			if f.isValid() {
				return f, submit(f)
			}
		}
	}

	return f.updateFields(msg)
}

func (f form) updateFields(msg tea.Msg) (form, tea.Cmd) {
	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)

	for i := range f.fields {
		f.fields[i], cmd = f.fields[i].Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return f, tea.Batch(cmds...)
}
