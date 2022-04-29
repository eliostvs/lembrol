package terminal

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/eliostvs/lembrol/internal/flashcard"
)

// MODEL

type deckStatus int

const (
	deckBrowsing deckStatus = iota
	deckCreating
	deckDeleting
	deckEditing
)

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
		edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
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
	edit    key.Binding
	delete  key.Binding
}

func newDecksModel(repository *flashcard.Repository, width, height int) decksModel {
	keys := newDeckKeys()
	delegate := list.NewDefaultDelegate()
	listModel := list.New(newDeckItems(repository.Deck.List()), &delegate, width, height)
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
			keys.edit,
			keys.delete,
			keys.study,
		}
	}

	return decksModel{
		delegate:   &delegate,
		keys:       keys,
		list:       listModel,
		repository: repository,
		width:      width,
		height:     height,
	}
}

type decksModel struct {
	form       form
	keys       *deckKeys
	list       list.Model
	repository *flashcard.Repository
	status     deckStatus
	delegate   *list.DefaultDelegate
	width      int
	height     int
}

// MESSAGE

type (
	initDeckMsg struct{}

	deletedDeckMsg struct {
		index int
	}

	editDeckMsg struct {
		index int
		item  deckItem
	}

	createdDeckMsg struct {
		index int
		item  deckItem
	}
)

// UPDATE

func (m decksModel) Init() tea.Cmd {
	return func() tea.Msg {
		return initDeckMsg{}
	}
}

// nolint:cyclop,gocyclo
func (m decksModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	currentDeck := toDeck(m.list)
	hasDeck := len(m.list.Items()) != 0

	resetControls := func() {
		m.delegate.Styles.SelectedTitle = selectedTitleStyle
		m.delegate.Styles.SelectedDesc = selectedDescStyle
		m.keys.add.SetEnabled(true)
		m.keys.confirm.SetEnabled(hasDeck)
		m.keys.confirm.SetHelp("enter", "open")
		m.keys.delete.SetEnabled(hasDeck)
		m.keys.edit.SetEnabled(hasDeck)
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
	case innerWindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height)
		return m, nil

	case initDeckMsg, canceledFormMsg:
		m.status = deckBrowsing
		resetControls()
		return m, nil

	case createdDeckMsg:
		m.list.InsertItem(msg.index, msg.item)
		m.status = deckBrowsing
		resetControls()
		return m, cmd

	case editDeckMsg:
		m.status = deckBrowsing
		m.list.RemoveItem(msg.index)
		m.list.InsertItem(msg.index-1, msg.item)
		resetControls()
		return m, nil

	case deletedDeckMsg:
		m.status = deckBrowsing
		m.list.RemoveItem(msg.index)
		resetControls()
		return m, nil

	case submittedFormMsg:
		if m.status == deckEditing {
			currentDeck.Name = m.form.Value("name")
			return m, editDeck(m.list.Index(), currentDeck, m.repository)
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

		case m.status == deckBrowsing && key.Matches(msg, m.keys.edit) && hasDeck:
			m.status = deckEditing
			m.form, cmd = createDeckForm(currentDeck.Name)
			return m, cmd

		case m.status == deckBrowsing && key.Matches(msg, m.keys.study) && currentDeck.HasDueCards():
			return m, startReview(currentDeck)

		case m.status == deckBrowsing && key.Matches(msg, m.keys.delete) && hasDeck:
			m.status = deckDeleting
			m.delegate.Styles.SelectedTitle = deletedTitle
			m.delegate.Styles.SelectedDesc = deletedDesc
			m.keys.confirm.SetHelp("enter", "confirm")
			m.keys.add.SetEnabled(false)
			m.keys.delete.SetEnabled(false)
			m.keys.edit.SetEnabled(false)
			m.keys.study.SetEnabled(false)
			m.list.NewStatusMessage(Red.Render("Delete this deck?"))
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
			return m, quit

		case m.status == deckBrowsing && key.Matches(msg, m.keys.confirm) && hasDeck:
			return m, showCards(currentDeck, firstCard)

		case m.status == deckDeleting && key.Matches(msg, m.keys.confirm):
			return m, deleteDeck(m.list.Index(), currentDeck, m.repository)

		case m.status == deckDeleting && key.Matches(msg, m.list.KeyMap.Quit):
			return m, m.Init()

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

func createDeckForm(name string) (form, tea.Cmd) {
	input := textinput.New()
	input.CharLimit = 30
	input.SetValue(name)
	input.CursorEnd()
	input.Prompt = inputPrompt
	input.TextStyle = Fuchsia
	input.PromptStyle = Fuchsia
	cmd := input.Focus()
	return newForm(newField("name", input, withLabel("Name"))), cmd
}

func createDeck(name string, repository *flashcard.Repository) tea.Cmd {
	return func() tea.Msg {
		deck, err := repository.Deck.Create(name)
		if err != nil {
			return fail(err)
		}
		return createdDeckMsg{index: 0, item: deckItem{deck}}
	}
}

func editDeck(index int, deck flashcard.Deck, repository *flashcard.Repository) tea.Cmd {
	return func() tea.Msg {
		if err := repository.Deck.Save(deck); err != nil {
			return fail(err)
		}
		return editDeckMsg{index: index, item: deckItem{deck}}
	}
}

func deleteDeck(index int, deck flashcard.Deck, repository *flashcard.Repository) tea.Cmd {
	return func() tea.Msg {
		if err := repository.Deck.Remove(deck); err != nil {
			return fail(err)
		}
		return deletedDeckMsg{index: index}
	}
}

// VIEW

func (m decksModel) View() string {
	switch m.status {
	case deckCreating:
		content := titleStyle.Render("New Deck")
		content += m.form.view()
		return largePaddingStyle.Render(content)

	case deckEditing:
		content := titleStyle.Render("Edit Deck")
		content += m.form.view()
		return largePaddingStyle.Render(content)

	case deckBrowsing, deckDeleting:
		fallthrough

	default:
		return midPaddingStyle.Render(m.list.View())
	}
}
