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

func (c *cursor) Last() {
	c.index = c.max
}

func (c *cursor) Max(max int) {
	if max >= 0 {
		c.max = max
		c.index = min(c.max, c.index)
	}
}

func (c *cursor) Update(msg tea.Msg) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyUp.String(), "k":
			c.Up()
		case tea.KeyDown.String(), "j":
			c.Down()
		}
	}
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
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

type FieldOption func(Field) Field

// WithMultiline enables multiline text input prepending the given prefix in the beginning of the new line.
func WithMultiline() FieldOption {
	return func(f Field) Field {
		f.multiline = true
		return f
	}
}

func WithLabel(label string) FieldOption {
	return func(f Field) Field {
		f.label = label
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
	label     string
	model     textinput.Model
	multiline bool
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

// SubmitForm triggers the form submit.
func SubmitForm(f Form) tea.Cmd {
	return func() tea.Msg {
		return submittedFormMsg{f}
	}
}

type submittedFormMsg struct {
	Form Form
}

// CancelForm triggers the form submit.
func CancelForm() tea.Cmd {
	return func() tea.Msg {
		return canceledFormMsg{}
	}
}

type canceledFormMsg struct {
}

// NewForm creates a new form with the given fields.
func NewForm(f Field, fields ...Field) Form {
	keys := newFormKeys()

	keys.next.SetEnabled(len(fields) != 0)
	keys.previous.SetEnabled(len(fields) != 0)

	return Form{
		cursor: newCursor(len(fields)),
		fields: append([]Field{f}, fields...),
		help:   help.NewModel(),
		keys:   keys,
	}
}

// Form is set of fields.
type Form struct {
	cursor cursor
	fields []Field
	help   help.Model
	keys   formKeys
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

func (f Form) view() string {
	var content string

	for _, field := range f.fields {
		content += field.View()
		content += "\n"
	}

	content += helpStyle.Render(f.help.View(f.keys))

	return formStyle.Render(content)
}
