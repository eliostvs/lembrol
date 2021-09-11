package terminal

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/muesli/reflow/truncate"

	"github.com/eliostvs/remembercli/internal/flashcard"
)

//go:embed templates
var files embed.FS

const encodedEnter = "¬"

var (
	encodedEnterMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(encodedEnter)}

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Dark: "#FFFDF5", Light: "#FFFDF5"}).
			Background(lipgloss.Color("#5A56E0")).
			Padding(0, 1).
			Bold(true)

	colors = map[string]lipgloss.Style{
		"fuchsia":     Fuchsia,
		"darkfuchsia": DarkFuchsia,
		"gray":        Gray,
		"lightgray":   LightGray,
		"green":       Green,
		"darkgreen":   DarkGreen,
		"indigo":      Indigo,
		"red":         Red,
		"darkred":     DarkRed,
		"white":       White,
	}

	funcMap = template.FuncMap{
		"title":       titleTag,
		"help":        helpTag,
		"style":       styleTag,
		"naturaltime": naturalTime,
		"markdown":    Markdown,
		"yesno":       yesno,
		"pluralize":   pluralize,
		"truncate":    truncateTag,
		"currentDeck": currentDeck,
		"hasDecks":    hasDecks,
		"hasDueCards": func(index int, decks []flashcard.Deck) bool {
			if len(decks) == 0 {
				return false
			}
			return decks[index].HasDueCards()
		},
	}
)

func titleTag(s string) string {
	return titleStyle.Render(s)
}

func naturalTime(t time.Time) string {
	if time.Now().Sub(t) < time.Minute {
		return "just now"
	}
	return humanize.Time(t)
}

func styleTag(name, s string) string {
	if style, ok := colors[name]; !ok {
		return s
	} else {
		return style.Render(s)
	}
}

func decodeMultiline(content, prefix string) string {
	return strings.Replace(content, encodedEnter, "\n"+prefix, -1)
}

func Markdown(width int, in string) (string, error) {
	r, _ := glamour.NewTermRenderer(
		glamour.WithEnvironmentConfig(),
		glamour.WithWordWrap(width),
	)

	lines, err := r.Render(decodeMultiline(in, ""))
	if err != nil {
		return "", err
	}

	return strings.ReplaceAll(lines, "\n", "\n "), nil
}

func yesno(val interface{}, yes, no string) string {
	if truth, _ := template.IsTrue(val); truth {
		return yes
	}
	return no
}

func pluralize(val int, suffix string) string {
	if val > 1 {
		return suffix
	}
	return ""
}

func truncateTag(width int, s string) string {
	return truncate.StringWithTail(s, uint(width-1), "…")
}

func helpTag(width int, sections ...string) string {
	if len(sections) == 0 {
		return ""
	}

	var lines []string
	for _, sections := range splitLines(width, sections) {
		lines = append(lines, renderLine(sections))
	}
	return strings.Join(lines, "\n")
}

func splitLines(width int, sections []string) [][]string {
	sectionDivisionWidth := 3
	var line, lineWidth int
	lines := [][]string{{}}

	for _, section := range sections {
		if len(section) == 0 {
			continue
		}

		sectionWidth := lipgloss.Width(section)

		if lineWidth+sectionWidth > width {
			line++
			lineWidth = 0
			lines = append(lines, []string{})
		}

		lines[line] = append(lines[line], section)
		lineWidth += sectionWidth + sectionDivisionWidth
	}

	return lines
}

func renderLine(sections []string) string {
	leftPadding := "   "

	b := strings.Builder{}
	for i, section := range sections {
		b.WriteString(section)
		if i < len(sections)-1 {
			b.WriteString(" • ")
		}
	}

	return leftPadding + b.String()
}

func newTemplates() *templates {
	files_, _ := fs.ReadDir(files, "templates")
	tmpls := make(map[string]*template.Template, len(files_))

	for _, file := range files_ {
		tmpls[withoutExt(file.Name())] = parse(file.Name())
	}

	return &templates{tmpls}
}

func parse(file string) *template.Template {
	return template.Must(
		template.New(file).
			Funcs(funcMap).
			ParseFS(files, filepath.Join("templates", file)),
	)
}

func withoutExt(name string) string {
	return name[:len(name)-len(filepath.Ext(name))]
}

type templates struct {
	templates map[string]*template.Template
}

func (t *templates) Render(name string, data interface{}) string {
	if _, ok := t.templates[name]; !ok {
		return fmt.Sprintf("template %s not found", name)
	}

	var buf bytes.Buffer
	if err := t.templates[name].Execute(&buf, data); err != nil {
		return fmt.Sprintf("%s: %v", name, err)
	}

	return "\n" + buf.String() + "\n"
}
