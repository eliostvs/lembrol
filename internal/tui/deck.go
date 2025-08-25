package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/eliostvs/lembrol/internal/flashcard"
)

// Messages

type (
	showBrowseDeckMsg struct {
		list list.Model
	}

	showAddDeckMsg struct {
		list list.Model
	}

	showEditDeckMsg struct {
		list list.Model
		deck flashcard.Deck
	}

	showDeleteDeckMsg struct {
		list list.Model
	}

	deckCreatedMsg struct {
		list list.Model
		deck flashcard.Deck
	}

	deckChangedMsg struct {
		list list.Model
		deck flashcard.Deck
	}

	deckDeletedMsg struct {
		list list.Model
	}
)

func showBrowseDeck(model list.Model) tea.Cmd {
	return func() tea.Msg {
		return showBrowseDeckMsg{list: model}
	}
}

func showAddDeck(model list.Model) tea.Cmd {
	return func() tea.Msg {
		return showAddDeckMsg{list: model}
	}
}

func showEditDeck(model list.Model, deck flashcard.Deck) tea.Cmd {
	return func() tea.Msg {
		return showEditDeckMsg{list: model, deck: deck}
	}
}

func showDeleteDeck(model list.Model) tea.Cmd {
	return func() tea.Msg {
		return showDeleteDeckMsg{list: model}
	}
}

func createDeck(name string, shared deckShared) tea.Cmd {
	return func() tea.Msg {
		deck, err := shared.repository.Create(name, nil)
		if err != nil {
			return fail(err)
		}

		return deckCreatedMsg{list: shared.list, deck: deck}
	}
}

func updateDeck(model list.Model, deck flashcard.Deck, repository Repository) tea.Cmd {
	return func() tea.Msg {
		if err := repository.Save(deck); err != nil {
			return fail(err)
		}

		return deckChangedMsg{list: model, deck: deck}
	}
}

func deleteDeck(model list.Model, deck flashcard.Deck, repository Repository) tea.Cmd {
	return func() tea.Msg {
		if err := repository.Delete(deck); err != nil {
			return fail(err)
		}
		return deckDeletedMsg{list: model}
	}
}

func hasDeck(m list.Model) bool {
	return len(m.Items()) > 0
}

func currentDeck(m list.Model) flashcard.Deck {
	item, ok := m.SelectedItem().(deckItem)
	if ok {
		return item.Deck
	}
	return flashcard.Deck{}
}

func hasDueCards(m list.Model) bool {
	return currentDeck(m).HasDueCards()
}

// Deck Item

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

// Browser Deck

type deckBrowseKeyMap struct {
	add    key.Binding
	open   key.Binding
	study  key.Binding
	edit   key.Binding
	delete key.Binding
}

func (k deckBrowseKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.add,
		k.open,
	}
}

func (k deckBrowseKeyMap) FullHelp() []key.Binding {
	return []key.Binding{
		k.add,
		k.open,
		k.edit,
		k.delete,
		k.study,
	}
}

func newDeckBrowsePage(shared deckShared) deckBrowsePage {
	shared.list.SetSize(shared.width-shared.styles.List.GetHorizontalFrameSize(), shared.height-shared.styles.List.GetVerticalFrameSize())
	shared.delegate.Styles.SelectedTitle = shared.styles.SelectedTitle
	shared.delegate.Styles.SelectedDesc = shared.styles.SelectedDesc
	return deckBrowsePage{
		deckShared: shared,
		keyMap: deckBrowseKeyMap{
			add: key.NewBinding(
				key.WithKeys("a"),
				key.WithHelp("a", "add"),
			),
			open: key.NewBinding(
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
				key.WithKeys("x", "delete"),
				key.WithHelp("x", "delete"),
			),
		},
	}.checkKeyMap()
}

type deckBrowsePage struct {
	deckShared
	keyMap deckBrowseKeyMap
}

func (m deckBrowsePage) Init() tea.Cmd {
	m.Log("deck-browse: Init")
	return nil
}

//nolint:cyclop
func (m deckBrowsePage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.Log(fmt.Sprintf("deck-browse: %T", msg))

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.keyMap.add):
			return m, showAddDeck(m.list)

		case key.Matches(msg, m.keyMap.edit):
			return m, showEditDeck(m.list, currentDeck(m.list))

		case key.Matches(msg, m.keyMap.study):
			return m, startReview(currentDeck(m.list))

		case key.Matches(msg, m.keyMap.delete):
			return m, showDeleteDeck(m.list)

		case key.Matches(msg, m.keyMap.open):
			return m, showCards(0, currentDeck(m.list))

		case key.Matches(msg, m.list.KeyMap.Quit) && m.list.FilterState() != list.FilterApplied:
			return m, quit
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m.checkKeyMap(), cmd
}

func (m deckBrowsePage) checkKeyMap() deckBrowsePage {
	hasDeck := hasDeck(m.list)
	m.keyMap.add.SetEnabled(m.list.FilterState() == list.Unfiltered)
	m.keyMap.open.SetEnabled(hasDeck)
	m.keyMap.delete.SetEnabled(hasDeck)
	m.keyMap.edit.SetEnabled(hasDeck)
	m.keyMap.study.SetEnabled(hasDueCards(m.list))
	m.list.NewStatusMessage("")
	m.list.SetFilteringEnabled(hasDeck)
	m.list.SetShowStatusBar(hasDeck)
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
	m.list.AdditionalShortHelpKeys = m.keyMap.ShortHelp
	m.list.AdditionalFullHelpKeys = m.keyMap.FullHelp
	return m
}

func (m deckBrowsePage) View() string {
	m.Log("deck-browse: View")
	return m.styles.List.Render(m.list.View())
}

// Form Deck

type deckFormKeyMap struct {
	confirm key.Binding
	cancel  key.Binding
}

func (k deckFormKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.confirm,
		k.cancel,
	}
}

func (k deckFormKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{}}
}

func newDeckForm(name string, shared Shared) deckForm {
	input := textinput.New()
	input.CharLimit = 30
	input.SetValue(name)
	input.CursorEnd()
	input.Focus()
	input.Prompt = "┃ "
	keyMap := deckFormKeyMap{
		confirm: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "confirm"),
		),
		cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}

	return deckForm{input: input, keyMap: keyMap, Shared: shared}
}

type deckForm struct {
	Shared
	input  textinput.Model
	keyMap deckFormKeyMap
}

func (m deckForm) Init() tea.Cmd {
	return m.input.Focus()
}

func (m deckForm) Update(msg tea.Msg) (deckForm, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.cancel):
			return m, cancelForm()

		case key.Matches(msg, m.keyMap.confirm):
			if m.isValid() {
				return m, submitForm(m.input)
			}
		}
	}

	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m deckForm) View() string {
	footer := lipgloss.
		NewStyle().
		Width(m.width).
		Padding(0, 2, 1).
		Render(renderHelp(m.keyMap, m.width, false))

	input := m.color().
		Height(m.height-lipgloss.Height(footer)).
		Width(m.width).
		Margin(0, 0, 1).
		Render(m.input.View())

	return lipgloss.JoinVertical(lipgloss.Top, input, footer)
}

func (m deckForm) color() lipgloss.Style {
	if m.isValid() {
		return lipgloss.NewStyle().Foreground(white)
	}
	return lipgloss.NewStyle().Foreground(red)
}

func (m deckForm) Value() string {
	return m.input.Value()
}

func (m deckForm) isValid() bool {
	return strings.TrimSpace(m.input.Value()) != ""
}

// Add Deck

func newDeckAddPage(shared deckShared) deckAddPage {
	return deckAddPage{form: newDeckForm("", shared.Shared), deckShared: shared}
}

type deckAddPage struct {
	deckShared
	form deckForm
}

func (m deckAddPage) Init() tea.Cmd {
	m.Log("deck-add: Init")
	return m.form.Init()
}

func (m deckAddPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.Log(fmt.Sprintf("deck-add: %T", msg))

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case submittedFormMsg[textinput.Model]:
		return m, tea.Batch(
			showLoading("Deck", "Creating deck..."),
			createDeck(msg.data.Value(), m.deckShared),
		)

	case canceledFormMsg:
		return m, showBrowseDeck(m.list)
	}

	m.form, cmd = m.form.Update(msg)
	return m, cmd
}

func (m deckAddPage) View() string {
	m.Log("deck-add: View")

	header := m.styles.Title.
		Margin(2, 0, 0, 2).
		Render("Decks")

	subTitle := m.styles.DimmedTitle.
		Margin(1, 0, 1, 2).
		Render("Add")

	m.form.height = m.height - lipgloss.Height(header) - lipgloss.Height(subTitle)
	form := m.styles.Text.Render(m.form.View())

	return lipgloss.JoinVertical(lipgloss.Top, header, subTitle, form)
}

// Edit Deck

func newEditDeckPage(deck flashcard.Deck, shared deckShared) deckEditPage {
	return deckEditPage{deck: deck, form: newDeckForm(deck.Name, shared.Shared), deckShared: shared}
}

type deckEditPage struct {
	deckShared
	deck flashcard.Deck
	form deckForm
}

func (m deckEditPage) Init() tea.Cmd {
	m.Log("deck-update: Init")
	return m.form.Init()
}

func (m deckEditPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.Log(fmt.Sprintf("deck-update: %T", msg))
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case submittedFormMsg[textinput.Model]:
		m.deck.Name = msg.data.Value()
		return m, tea.Batch(
			showLoading("Deck", "Saving deck..."),
			updateDeck(m.list, m.deck, m.repository),
		)

	case canceledFormMsg:
		return m, showBrowseDeck(m.list)
	}

	m.form, cmd = m.form.Update(msg)
	return m, cmd
}

func (m deckEditPage) View() string {
	m.Log("deck-update: View")
	header := m.styles.Title.
		Margin(2, 0, 0, 2).
		Render("Decks")

	subTitle := m.styles.DimmedTitle.
		Margin(1, 0, 1, 2).
		Render("Edit")

	m.form.height = m.height - lipgloss.Height(header) - lipgloss.Height(subTitle)
	form := m.styles.Text.Render(m.form.View())

	return lipgloss.JoinVertical(lipgloss.Top, header, subTitle, form)
}

// Delete Deck

type deckDeleteKeyMap struct {
	confirm key.Binding
	cancel  key.Binding
}

func (k deckDeleteKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.confirm,
		k.cancel,
	}
}

func (k deckDeleteKeyMap) FullHelp() []key.Binding {
	return []key.Binding{
		k.confirm,
		k.cancel,
	}
}

func newDeleteDeckPage(shared deckShared) deckDeletePage {
	keyMap := deckDeleteKeyMap{
		cancel: key.NewBinding(
			key.WithKeys(tea.KeyEsc.String()),
			key.WithHelp("esc", "cancel"),
		),
		confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
	}

	shared.list.SetSize(shared.width-shared.styles.List.GetHorizontalFrameSize(), shared.height-shared.styles.List.GetVerticalFrameSize())
	shared.delegate.Styles.SelectedTitle = shared.styles.DeletedTitle
	shared.delegate.Styles.SelectedDesc = shared.styles.DeletedDesc

	shared.list.AdditionalShortHelpKeys = keyMap.ShortHelp
	shared.list.AdditionalFullHelpKeys = keyMap.FullHelp
	shared.list.KeyMap.CloseFullHelp.SetEnabled(false)
	shared.list.KeyMap.CursorDown.SetEnabled(false)
	shared.list.KeyMap.CursorUp.SetEnabled(false)
	shared.list.KeyMap.Filter.SetEnabled(false)
	shared.list.KeyMap.GoToEnd.SetEnabled(false)
	shared.list.KeyMap.GoToStart.SetEnabled(false)
	shared.list.KeyMap.NextPage.SetEnabled(false)
	shared.list.KeyMap.PrevPage.SetEnabled(false)
	shared.list.KeyMap.ShowFullHelp.SetEnabled(false)
	shared.list.KeyMap.Quit.SetEnabled(false)
	shared.list.NewStatusMessage(shared.styles.DeletedStatus.Render("Delete this deck?"))

	return deckDeletePage{
		deckShared: shared,
		keyMap:     keyMap,
	}
}

type deckDeletePage struct {
	deckShared
	keyMap deckDeleteKeyMap
}

func (m deckDeletePage) Init() tea.Cmd {
	m.Log("deck-delete: Init")
	return nil
}

func (m deckDeletePage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.Log(fmt.Sprintf("deck-delete: %T", msg))

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.confirm):
			return m, tea.Batch(
				showLoading("Decks", "Deleting deck..."),
				deleteDeck(m.list, currentDeck(m.list), m.repository),
			)

		case key.Matches(msg, m.keyMap.cancel):
			return m, showBrowseDeck(m.list)
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m deckDeletePage) View() string {
	m.Log("deck-delete: View")
	return m.styles.List.Render(m.list.View())
}

// Deck Page

type deckShared struct {
	Shared
	list     list.Model
	delegate *list.DefaultDelegate
}

func newDeckPage(parent Shared, index int) deckPage {
	delegate := list.NewDefaultDelegate()
	shared := deckShared{
		Shared:   parent,
		delegate: &delegate,
		list:     list.New(newDeckItems(parent.repository.List()), &delegate, parent.width, parent.height),
	}
	shared.list.SetSize(parent.width, parent.height)
	shared.list.Select(index)
	shared.list.Title = "Decks"
	shared.list.Styles.NoItems = shared.list.Styles.NoItems.Copy().Margin(0, 2)

	return deckPage{
		deckShared: shared,
		page:       newDeckBrowsePage(shared),
	}
}

type deckPage struct {
	deckShared
	page tea.Model
}

func (m deckPage) Init() tea.Cmd {
	m.Log("deck: Init")
	return m.page.Init()
}

func (m deckPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.Log(fmt.Sprintf("deck: %T", msg))

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case showLoadingMsg:
		m.page = newLoadingPage(m.Shared, msg.title, msg.description)
		return m, m.page.Init()

	case showBrowseDeckMsg:
		m.list = msg.list
		m.page = newDeckBrowsePage(m.deckShared)
		return m, nil

	case showAddDeckMsg:
		m.list = msg.list
		m.page = newDeckAddPage(m.deckShared)
		return m, m.page.Init()

	case showEditDeckMsg:
		m.list = msg.list
		m.page = newEditDeckPage(msg.deck, m.deckShared)
		return m, m.page.Init()

	case showDeleteDeckMsg:
		m.list = msg.list
		m.page = newDeleteDeckPage(m.deckShared)
		return m, cmd

	case deckCreatedMsg:
		m.list = msg.list
		m.list.InsertItem(m.list.Index(), deckItem{msg.deck})
		m.list.ResetFilter()
		m.page = newDeckBrowsePage(m.deckShared)
		return m, nil

	case deckChangedMsg:
		m.list = msg.list
		m.list.RemoveItem(m.list.Index())
		m.list.InsertItem(m.list.Index()-1, deckItem{msg.deck})
		m.list.ResetFilter()
		m.page = newDeckBrowsePage(m.deckShared)
		return m, nil

	case deckDeletedMsg:
		m.list = msg.list
		m.list.RemoveItem(m.list.Index())
		m.list.ResetFilter()
		m.page = newDeckBrowsePage(m.deckShared)
		return m, nil
	}

	m.page, cmd = m.page.Update(msg)
	return m, cmd
}

func (m deckPage) View() string {
	m.Log("deck: View")
	return m.page.View()
}
