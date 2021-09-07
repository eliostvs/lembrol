package terminal

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type FieldOption func(Field) Field

// WithMultiline enables multiline text input prepending the given prefix in the beginning of the new line.
func WithMultiline(prefix string) FieldOption {
	return func(f Field) Field {
		f.multiline = true
		f.prefix = prefix
		return f
	}
}

// NewField creates a new field.
func NewField(name string, model textinput.Model, options ...FieldOption) Field {
	f := Field{name: name, model: model}
	for _, option := range options {
		f = option(f)
	}
	return f
}

// Field specifies an input Field where the user can enter data.
type Field struct {
	name      string
	model     textinput.Model
	multiline bool
	prefix    string
}

// Focus sets the focus on this field.
func (f Field) Focus() (Field, tea.Cmd) {
	f.model.PromptStyle = Fuchsia
	f.model.TextStyle = Fuchsia
	return f, f.model.Focus()
}

// Blur removes the focus from this field.
func (f Field) Blur() Field {
	f.model.PromptStyle = DarkFuchsia
	f.model.TextStyle = DarkFuchsia
	f.model.Blur()
	return f
}

// Update changes the input model.
func (f Field) Update(msg tea.Msg) (Field, tea.Cmd) {
	var cmd tea.Cmd
	f.model, cmd = f.model.Update(msg)
	return f, cmd
}

// Match returns if the input has the given name.
func (f Field) Match(name string) bool {
	return f.name == name
}

// IsValid returns is input content is valid.
func (f Field) IsValid() bool {
	return 0 < len(f.model.Value())
}

// View renders the input.
func (f Field) View() string {
	return decodeMultiline(f.model.View(), f.prefix)
}

// Focused returns if the input is focused.
func (f Field) Focused() bool {
	return f.model.Focused()
}

// BreakLine inserts a new line.
func (f Field) BreakLine() Field {
	if !f.multiline {
		return f
	}
	f.model, _ = f.model.Update(encodedEnterMsg)
	return f
}

// Value returns the input value.
func (f Field) Value() string {
	return f.model.Value()
}

// Submit triggers the form submit.
func Submit(f Form) tea.Cmd {
	return func() tea.Msg {
		return submittedFormMsg{f}
	}
}

type submittedFormMsg struct {
	Form Form
}

// NewForm creates a new form with the given fields.
func NewForm(f Field, fields ...Field) Form {
	return Form{
		cursor: newCursor(len(fields)),
		fields: append([]Field{f}, fields...),
	}
}

// Form is set of fields.
type Form struct {
	cursor cursor
	fields []Field
}

func (f Form) focus(index int) (Form, tea.Cmd) {
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

// Width sets the width of all fields.
func (f Form) Width(width int) Form {
	for i := range f.fields {
		f.fields[i].model.Width = width
	}

	return f
}

// Error returns if the given field has error.
func (f Form) Error(name string) bool {
	for _, field := range f.fields {
		if field.Match(name) {
			return !field.IsValid()
		}
	}
	return false
}

func (f Form) isValid() bool {
	for _, field := range f.fields {
		if !field.IsValid() {
			return false
		}
	}
	return true
}

// Value returns the given field value.
func (f Form) Value(name string) string {
	for _, field := range f.fields {
		if field.Match(name) {
			return field.Value()
		}
	}
	return ""
}

// View returns the given field view.
func (f Form) View(name string) string {
	for _, field := range f.fields {
		if field.Match(name) {
			return field.View()
		}
	}
	return ""
}

func (f Form) prev() (Form, tea.Cmd) {
	f.cursor.Up()
	return f.focus(f.cursor.Value())
}

func (f Form) next() (Form, tea.Cmd) {
	f.cursor.Down()
	return f.focus(f.cursor.Value())
}

// Update the form fields inner state.
func (f Form) Update(msg tea.Msg) (Form, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "shift+tab", tea.KeyUp.String():
			return f.prev()

		case "tab", tea.KeyDown.String():
			return f.next()

		case "alt+enter":
			for i := range f.fields {
				if f.fields[i].Focused() {
					f.fields[i] = f.fields[i].BreakLine()
					return f, nil
				}
			}

		case tea.KeyEnter.String():
			if f.isValid() {
				return f, Submit(f)
			}
		}
	}

	return f.updateFields(msg)
}

func (f Form) updateFields(msg tea.Msg) (Form, tea.Cmd) {
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
