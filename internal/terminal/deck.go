package terminal

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/eliostvs/remembercli/internal/flashcard"
)

// STATE

type deckStatus int

const (
	deckBrowsing deckStatus = iota
	deckCreating
	deckDeleting
	deckEditing
)

// ITEM

type deckItem struct {
	flashcard.Deck
}

func (d deckItem) Title() string {
	return d.Name
}

func (d deckItem) Description() string {
	return fmt.Sprintf(
		"%d card%s | %d due",
		d.Total(),
		pluralize(d.Total(), "s"),
		len(d.DueCards()),
	)
}

func (d deckItem) FilterValue() string {
	return d.Name
}

func newDeckItems(decks []flashcard.Deck) []list.Item {
	items := make([]list.Item, 0, len(decks))
	for _, deck := range decks {
		items = append(items, deckItem{deck})
	}
	return items
}

// KEYS

func newDeckKeys() *deckKeys {
	return &deckKeys{
		add: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add"),
		),
		confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "open"),
		),
		study: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "study"),
		),
		rename: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "rename"),
		),
		delete: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "delete"),
		),
	}
}

type deckKeys struct {
	add     key.Binding
	confirm key.Binding
	study   key.Binding
	rename  key.Binding
	delete  key.Binding
}

// MODEL

func newDecksModel(decks []flashcard.Deck, repo *flashcard.Repository) decksModel {
	keys := newDeckKeys()
	delegate := list.NewDefaultDelegate()
	listModel := list.NewModel(newDeckItems(decks), &delegate, 0, 0)
	listModel.Title = "Decks"
	listModel.Styles.Title = titleStyle
	listModel.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			keys.add,
			keys.confirm,
		}
	}
	listModel.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			keys.add,
			keys.confirm,
			keys.rename,
			keys.delete,
			keys.study,
		}
	}
	listModel.DisableQuitKeybindings()

	return decksModel{
		delegate:   &delegate,
		keys:       keys,
		list:       listModel,
		repository: repo,
	}
}

type decksModel struct {
	form       Form
	keys       *deckKeys
	list       list.Model
	repository *flashcard.Repository
	status     deckStatus
	delegate   *list.DefaultDelegate
	viewport   viewport
}

// VIEW

func (m decksModel) View() string {

	switch m.status {
	case deckCreating:
		content := titleStyle.Render("New Deck")
		content += m.form.view()
		return largePaddingStyle.Render(content)

	case deckEditing:
		content := titleStyle.Render("Rename Deck")
		content += m.form.view()
		return largePaddingStyle.Render(content)

	case deckBrowsing, deckDeleting:
		fallthrough

	default:
		return midPaddingStyle.Render(m.list.View())
	}
}

// INIT

func (m decksModel) init() tea.Cmd {
	return func() tea.Msg {
		return initDeckMsg{}
	}
}

// UPDATE

type (
	initDeckMsg struct{}

	deletedDeckMsg struct {
		index int
	}

	renamedDeckMsg struct {
		index int
		item  deckItem
	}

	createdDeckMsg struct {
		index int
		item  deckItem
	}
)

// nolint:cyclop,gocyclo
func (m decksModel) Update(msg tea.Msg) (decksModel, tea.Cmd) {
	var cmd tea.Cmd

	m.list.SetHeight(m.viewport.height)
	m.list.SetWidth(m.viewport.width)

	currentDeck := toDeck(m.list)
	hasDeck := len(m.list.Items()) != 0

	resetControls := func() {
		m.delegate.Styles.SelectedTitle = selectedTitleStyle
		m.delegate.Styles.SelectedDesc = selectedDescStyle
		m.keys.add.SetEnabled(true)
		m.keys.confirm.SetEnabled(hasDeck)
		m.keys.confirm.SetHelp("enter", "open")
		m.keys.delete.SetEnabled(hasDeck)
		m.keys.rename.SetEnabled(hasDeck)
		m.keys.study.SetEnabled(hasDeck)
		m.list.NewStatusMessage("")
		m.list.SetFilteringEnabled(hasDeck)
		m.list.KeyMap.CloseFullHelp.SetEnabled(true)
		m.list.KeyMap.CursorDown.SetEnabled(hasDeck)
		m.list.KeyMap.CursorDown.SetEnabled(hasDeck)
		m.list.KeyMap.CursorUp.SetEnabled(hasDeck)
		m.list.KeyMap.CursorUp.SetEnabled(hasDeck)
		m.list.KeyMap.Filter.SetEnabled(hasDeck)
		m.list.KeyMap.GoToEnd.SetEnabled(hasDeck)
		m.list.KeyMap.GoToStart.SetEnabled(hasDeck)
		m.list.KeyMap.NextPage.SetEnabled(hasDeck)
		m.list.KeyMap.PrevPage.SetEnabled(hasDeck)
		m.list.KeyMap.ShowFullHelp.SetEnabled(true)
	}

	switch msg := msg.(type) {
	case initDeckMsg, canceledFormMsg:
		m.status = deckBrowsing
		resetControls()
		return m, nil

	case createdDeckMsg:
		m.list.InsertItem(msg.index, msg.item)
		return m, m.init()

	case renamedDeckMsg:
		m.list.RemoveItem(msg.index)
		m.list.InsertItem(msg.index-1, msg.item)
		return m, m.init()

	case deletedDeckMsg:
		m.list.RemoveItem(msg.index)
		return m, m.init()

	case submittedFormMsg:
		if m.status == deckEditing {
			currentDeck.Name = m.form.Value("name")
			return m, renameDeck(m.list.Index(), currentDeck, m.repository)
		}

		if m.status == deckCreating {
			return m, createDeck(m.form.Value("name"), m.repository)
		}

	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case m.status == deckBrowsing && key.Matches(msg, m.keys.add):
			m.status = deckCreating
			m.form, cmd = createDeckForm("")
			return m, cmd

		case m.status == deckBrowsing && key.Matches(msg, m.keys.rename) && hasDeck:
			m.status = deckEditing
			m.form, cmd = createDeckForm(currentDeck.Name)
			return m, cmd

		case m.status == deckBrowsing && key.Matches(msg, m.keys.study) && currentDeck.HasDueCards():
			return m, startReview(currentDeck)

		case m.status == deckBrowsing && key.Matches(msg, m.keys.delete) && hasDeck:
			m.status = deckDeleting
			m.delegate.Styles.SelectedTitle = deletedTitle
			m.delegate.Styles.SelectedDesc = deletedDesc
			m.list.NewStatusMessage(Red.Render("Delete this deck?"))
			m.keys.confirm.SetHelp("enter", "confirm")
			m.keys.add.SetEnabled(false)
			m.keys.delete.SetEnabled(false)
			m.keys.rename.SetEnabled(false)
			m.keys.study.SetEnabled(false)
			m.list.KeyMap.CloseFullHelp.SetEnabled(false)
			m.list.KeyMap.CursorDown.SetEnabled(false)
			m.list.KeyMap.CursorUp.SetEnabled(false)
			m.list.KeyMap.Filter.SetEnabled(false)
			m.list.KeyMap.GoToEnd.SetEnabled(false)
			m.list.KeyMap.GoToStart.SetEnabled(false)
			m.list.KeyMap.NextPage.SetEnabled(false)
			m.list.KeyMap.PrevPage.SetEnabled(false)
			m.list.KeyMap.ShowFullHelp.SetEnabled(false)
			return m, nil

		case m.status == deckBrowsing && key.Matches(msg, m.list.KeyMap.Quit) && m.list.FilterState() != list.FilterApplied:
			return m, exitCmd

		case m.status == deckBrowsing && key.Matches(msg, m.keys.confirm) && hasDeck:
			return m, showCards(currentDeck)

		case m.status == deckDeleting && key.Matches(msg, m.keys.confirm):
			return m, deleteDeck(m.list.Index(), currentDeck, m.repository)

		case m.status == deckDeleting && key.Matches(msg, m.list.KeyMap.Quit):
			return m, m.init()

			// the only two actions in delete state should confirm or cancel
		case m.status == deckDeleting:
			return m, nil
		}
	}

	if m.status == deckEditing || m.status == deckCreating {
		m.form, cmd = m.form.Update(msg)
		return m, cmd
	}

	m.list, cmd = m.list.Update(msg)
	resetControls()
	return m, cmd
}

func toDeck(l list.Model) flashcard.Deck {
	item, ok := l.SelectedItem().(deckItem)
	if ok {
		return item.Deck
	}
	return flashcard.Deck{}
}

func createDeckForm(name string) (Form, tea.Cmd) {
	input := textinput.NewModel()
	input.CharLimit = 30
	input.SetValue(name)
	input.CursorEnd()
	input.Prompt = inputPrompt
	input.TextStyle = Fuchsia
	input.PromptStyle = Fuchsia
	cmd := input.Focus()
	return NewForm(NewField("name", input, WithLabel("Name"))), cmd
}

func createDeck(name string, repo *flashcard.Repository) tea.Cmd {
	return func() tea.Msg {
		deck, err := repo.Create(name)
		if err != nil {
			return failed(err)
		}
		return createdDeckMsg{index: 0, item: deckItem{deck}}
	}
}

func renameDeck(index int, deck flashcard.Deck, repo *flashcard.Repository) tea.Cmd {
	return func() tea.Msg {
		if err := repo.Save(deck); err != nil {
			return failed(err)
		}
		return renamedDeckMsg{index: index, item: deckItem{deck}}
	}
}

func deleteDeck(index int, deck flashcard.Deck, repo *flashcard.Repository) tea.Cmd {
	return func() tea.Msg {
		if err := repo.Remove(deck); err != nil {
			return failed(err)
		}
		return deletedDeckMsg{index: index}
	}
}
