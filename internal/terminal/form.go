package terminal

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func newCursor(max int) cursor {
	return cursor{max: max}
}

type cursor struct {
	index int
	max   int
}

func (c *cursor) Up() {
	if c.index > 0 {
		c.index--
	}
}

func (c *cursor) Down() {
	if c.index < c.max {
		c.index++
	}
}

func (c *cursor) Value() int {
	return c.index
}

var encodedEnterMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(encodedEnter)}

func newFormKeys() formKeys {
	return formKeys{
		submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		previous: key.NewBinding(
			key.WithKeys("up", "shift+tab"),
			key.WithHelp("↑", "up"),
		),
		next: key.NewBinding(
			key.WithKeys("down", "tab"),
			key.WithHelp("↓", "down"),
		),
		newline: key.NewBinding(
			key.WithKeys("alt+enter"),
			key.WithHelp("alt+enter", "new line"),
		),
	}
}

type formKeys struct {
	submit   key.Binding
	cancel   key.Binding
	previous key.Binding
	next     key.Binding
	newline  key.Binding
}

func (k formKeys) ShortHelp() []key.Binding {
	return []key.Binding{
		k.next,
		k.previous,
		k.submit,
		k.cancel,
	}
}

func (k formKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{{}}
}

type fieldOption func(field) field

// withMultiline enables multiline text input prepending the given prefix in the beginning of the new line.
func withMultiline() fieldOption {
	return func(f field) field {
		f.multiline = true
		return f
	}
}

func withLabel(label string) fieldOption {
	return func(f field) field {
		f.label = label
		return f
	}
}

// newField creates a new field.
func newField(name string, model textinput.Model, options ...fieldOption) field {
	f := field{name: name, model: model}
	for _, option := range options {
		f = option(f)
	}
	return f
}

// field specifies an input field where the user can enter data.
type field struct {
	name      string
	label     string
	model     textinput.Model
	multiline bool
}

// Focus sets the focus on this field.
func (f field) Focus() (field, tea.Cmd) {
	f.model.PromptStyle = Fuchsia
	f.model.TextStyle = Fuchsia
	return f, f.model.Focus()
}

// Blur removes the focus from this field.
func (f field) Blur() field {
	f.model.PromptStyle = DarkFuchsia
	f.model.TextStyle = DarkFuchsia
	f.model.Blur()
	return f
}

// Update changes the input model.
func (f field) Update(msg tea.Msg) (field, tea.Cmd) {
	var cmd tea.Cmd
	f.model, cmd = f.model.Update(msg)
	return f, cmd
}

// Match returns if the input has the given name.
func (f field) Match(name string) bool {
	return f.name == name
}

// IsValid returns is input content is valid.
func (f field) IsValid() bool {
	return 0 < len(f.model.Value())
}

// View renders the input.
func (f field) View() string {
	color := White
	if !f.IsValid() {
		color = Red
	}

	var content string

	if f.label != "" {
		content += color.Render(f.label)
		content += "\n"
	}

	content += renderMultiline(f.model.View(), f.model.Prompt)
	return fieldStyle.Render(color.Render(content))
}

// Focused returns if the input is focused.
func (f field) Focused() bool {
	return f.model.Focused()
}

// BreakLine inserts a new line.
func (f field) BreakLine() field {
	if !f.multiline {
		return f
	}
	f.model, _ = f.model.Update(encodedEnterMsg)
	return f
}

// Value returns the input value.
func (f field) Value() string {
	return f.model.Value()
}

// SubmitForm triggers the form submit.
func SubmitForm(f form) tea.Cmd {
	return func() tea.Msg {
		return submittedFormMsg{f}
	}
}

type submittedFormMsg struct {
	Form form
}

// CancelForm triggers the form submit.
func CancelForm() tea.Cmd {
	return func() tea.Msg {
		return canceledFormMsg{}
	}
}

type canceledFormMsg struct {
}

// newForm creates a new form with the given fields.
func newForm(f field, fields ...field) form {
	keys := newFormKeys()
	keys.next.SetEnabled(len(fields) != 0)
	keys.previous.SetEnabled(len(fields) != 0)
	return form{
		cursor: newCursor(len(fields)),
		fields: append([]field{f}, fields...),
		help:   help.New(),
		keys:   keys,
	}
}

// form is set of fields.
type form struct {
	cursor cursor
	fields []field
	help   help.Model
	keys   formKeys
}

func (f form) focus(index int) (form, tea.Cmd) {
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

func (f form) isValid() bool {
	for _, field := range f.fields {
		if !field.IsValid() {
			return false
		}
	}
	return true
}

// Value returns the given field value.
func (f form) Value(name string) string {
	for _, field := range f.fields {
		if field.Match(name) {
			return field.Value()
		}
	}
	return ""
}

func (f form) prev() (form, tea.Cmd) {
	f.cursor.Up()
	return f.focus(f.cursor.Value())
}

func (f form) next() (form, tea.Cmd) {
	f.cursor.Down()
	return f.focus(f.cursor.Value())
}

// Update the form fields inner state.
func (f form) Update(msg tea.Msg) (form, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, f.keys.previous):
			return f.prev()

		case key.Matches(msg, f.keys.next):
			return f.next()

		case key.Matches(msg, f.keys.newline):
			for i := range f.fields {
				if f.fields[i].Focused() {
					f.fields[i] = f.fields[i].BreakLine()
				}
			}
			return f, nil

		case key.Matches(msg, f.keys.cancel):
			return f, CancelForm()

		case key.Matches(msg, f.keys.submit):
			if f.isValid() {
				return f, SubmitForm(f)
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

func (f form) view() string {
	var content string

	for _, field := range f.fields {
		content += field.View()
		content += "\n"
	}

	content += helpStyle.Render(f.help.View(f.keys))

	return formStyle.Render(content)
}
