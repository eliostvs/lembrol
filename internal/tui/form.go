package tui

import (
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

func submitForm[T any](data T) tea.Cmd {
	return func() tea.Msg {
		return submittedFormMsg[T]{data}
	}
}

type submittedFormMsg[T any] struct {
	data T
}

func cancelForm() tea.Cmd {
	return func() tea.Msg {
		return canceledFormMsg{}
	}
}

type canceledFormMsg struct{}
