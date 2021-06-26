package terminal

import (
	"reflect"

	"github.com/charmbracelet/bubbles/paginator"
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

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

func newPosition(items int) position {
	page := paginator.NewModel()
	page.Type = paginator.Dots
	page.PerPage = 5
	page.ActiveDot = Fuchsia.Render("•")
	page.InactiveDot = Gray.Render("•")
	page.SetTotalPages(items)
	return position{paginator: page, cursor: newCursor(page.ItemsOnPage(items)), items: items}
}

type position struct {
	cursor    cursor
	items     int
	paginator paginator.Model
}

func (p position) HasFocus(i int) bool {
	return p.cursor.Value() == i
}

func (p position) Paginator() string {
	if p.paginator.TotalPages > 1 {
		return p.paginator.View()
	}
	return ""
}

func (p position) Update(msg tea.KeyMsg) position {
	p.paginator, _ = p.paginator.Update(msg)
	p.cursor.Max(p.paginator.ItemsOnPage(p.items) - 1)
	p.cursor.Update(msg)
	return p
}

func (p position) Item() int {
	cursor := p.cursor.Value()
	page := p.paginator.Page
	perPage := p.paginator.PerPage
	return cursor + perPage*page
}

func (p position) Items(items interface{}) interface{} {
	val := reflect.ValueOf(items)
	total := val.Len()
	start, end := p.paginator.GetSliceBounds(total)
	return val.Slice(start, end).Interface()
}

func (p position) Decrease() position {
	p.items--
	p.paginator.SetTotalPages(p.items)
	if p.paginator.Page > 0 && p.items%p.paginator.PerPage == 0 {
		p.paginator.PrevPage()
	}
	p.cursor.Max(p.paginator.ItemsOnPage(p.items) - 1)
	return p
}

func (p position) Increase() position {
	p.items++
	p.paginator.SetTotalPages(p.items)
	p.paginator.Page = max(0, p.paginator.TotalPages-1)
	p.cursor.Last()
	p.cursor.Max(p.paginator.ItemsOnPage(p.items) - 1)
	return p
}
