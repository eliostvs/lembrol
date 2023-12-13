package terminal

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

func createDeck(name string, common deckCommon) tea.Cmd {
	return func() tea.Msg {
		deck, err := common.repository.Create(name, nil)
		if err != nil {
			return fail(err)
		}

		return deckCreatedMsg{list: common.list, deck: deck}
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

func newDeckBrowsePage(common deckCommon) deckBrowsePage {
	common.list.SetSize(common.width, common.height)
	common.delegate.Styles.SelectedTitle = selectedTitleStyle
	common.delegate.Styles.SelectedDesc = selectedDescStyle
	return deckBrowsePage{
		deckCommon: common,
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
	deckCommon
	keyMap deckBrowseKeyMap
}

func (m deckBrowsePage) Init() tea.Cmd {
	return nil
}

//nolint:cyclop
func (m deckBrowsePage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case innerWindowSizeMsg:
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
	return midPaddingStyle.Render(m.list.View())
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

func newDeckForm(name string, common appCommon) (deckForm, tea.Cmd) {
	input := textinput.New()
	input.CharLimit = 30
	input.SetValue(name)
	input.CursorEnd()
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

	return deckForm{input: input, keyMap: keyMap, appCommon: common}, input.Focus()
}

type deckForm struct {
	appCommon
	input  textinput.Model
	keyMap deckFormKeyMap
}

func (m deckForm) Init() tea.Cmd {
	return nil
}

func (m deckForm) Update(msg tea.Msg) (deckForm, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case innerWindowSizeMsg:
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
	var content strings.Builder
	content.WriteString(m.color().Copy().Padding(2, 0).Render(m.input.View()))
	content.WriteString(renderHelp(m.keyMap, m.width, m.height-lipgloss.Height(content.String()), false))
	return content.String()
}

func (m deckForm) color() lipgloss.Style {
	if m.isValid() {
		return White
	}
	return Red
}

func (m deckForm) Value() string {
	return m.input.Value()
}

func (m deckForm) isValid() bool {
	return strings.TrimSpace(m.input.Value()) != ""
}

// Add Deck

func newDeckAddPage(common deckCommon) (deckAddPage, tea.Cmd) {
	form, cmd := newDeckForm("", common.appCommon)
	return deckAddPage{form: form, deckCommon: common}, cmd
}

type deckAddPage struct {
	deckCommon
	form deckForm
}

func (m deckAddPage) Init() tea.Cmd {
	return nil
}

func (m deckAddPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case submittedFormMsg[textinput.Model]:
		return m, tea.Batch(
			showLoading("Deck", "Creating deck..."),
			createDeck(msg.data.Value(), m.deckCommon),
		)

	case canceledFormMsg:
		return m, showBrowseDeck(m.list)
	}

	m.form, cmd = m.form.Update(msg)
	return m, cmd
}

func (m deckAddPage) View() string {
	var content strings.Builder
	content.WriteString(titleStyle.Render("New Deck"))
	content.WriteString(m.form.View())
	return largePaddingStyle.Render(content.String())
}

// Edit Deck

func newEditDeckPage(deck flashcard.Deck, common deckCommon) (deckEditPage, tea.Cmd) {
	form, cmd := newDeckForm(deck.Name, common.appCommon)
	return deckEditPage{deck: deck, form: form, deckCommon: common}, cmd
}

type deckEditPage struct {
	deckCommon
	deck flashcard.Deck
	form deckForm
}

func (m deckEditPage) Init() tea.Cmd {
	return nil
}

func (m deckEditPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
	var content strings.Builder
	content.WriteString(titleStyle.Render("Edit Deck"))
	content.WriteString(m.form.View())
	return largePaddingStyle.Render(content.String())
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

func newDeleteDeckPage(common deckCommon) deckDeletePage {
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

	common.list.SetSize(common.width, common.height)
	common.delegate.Styles.SelectedTitle = deletedTitleStyle
	common.delegate.Styles.SelectedDesc = deletedDescStyle

	common.list.AdditionalShortHelpKeys = keyMap.ShortHelp
	common.list.AdditionalFullHelpKeys = keyMap.FullHelp
	common.list.KeyMap.CloseFullHelp.SetEnabled(false)
	common.list.KeyMap.CursorDown.SetEnabled(false)
	common.list.KeyMap.CursorUp.SetEnabled(false)
	common.list.KeyMap.Filter.SetEnabled(false)
	common.list.KeyMap.GoToEnd.SetEnabled(false)
	common.list.KeyMap.GoToStart.SetEnabled(false)
	common.list.KeyMap.NextPage.SetEnabled(false)
	common.list.KeyMap.PrevPage.SetEnabled(false)
	common.list.KeyMap.ShowFullHelp.SetEnabled(false)
	common.list.KeyMap.Quit.SetEnabled(false)
	common.list.NewStatusMessage(Red.Render("Delete this deck?"))

	return deckDeletePage{
		deckCommon: common,
		keyMap:     keyMap,
	}
}

type deckDeletePage struct {
	deckCommon
	keyMap deckDeleteKeyMap
}

func (m deckDeletePage) Init() tea.Cmd {
	return nil
}

func (m deckDeletePage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case innerWindowSizeMsg:
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
	return midPaddingStyle.Render(m.list.View())
}

// Deck Page

type deckCommon struct {
	appCommon
	list     list.Model
	delegate *list.DefaultDelegate
}

func newDeckPage(index int, parent appCommon) deckPage {
	delegate := list.NewDefaultDelegate()
	common := deckCommon{
		appCommon: parent,
		delegate:  &delegate,
		list:      list.New(newDeckItems(parent.repository.List()), &delegate, parent.width, parent.height),
	}
	common.list.SetSize(parent.width, parent.height)
	common.list.Select(index)
	common.list.Title = "Decks"
	common.list.Styles.NoItems = common.list.Styles.NoItems.Copy().Margin(0, 2)

	return deckPage{
		deckCommon: common,
		page:       newDeckBrowsePage(common),
	}
}

type deckPage struct {
	deckCommon
	page tea.Model
}

func (m deckPage) Init() tea.Cmd {
	return m.page.Init()
}

func (m deckPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case showLoadingMsg:
		m.page = newLoadingPage(msg.title, msg.description, m.appCommon)
		return m, m.page.Init()

	case showBrowseDeckMsg:
		m.list = msg.list
		m.page = newDeckBrowsePage(m.deckCommon)
		return m, nil

	case showAddDeckMsg:
		m.list = msg.list
		m.page, cmd = newDeckAddPage(m.deckCommon)
		return m, cmd

	case showEditDeckMsg:
		m.list = msg.list
		m.page, cmd = newEditDeckPage(msg.deck, m.deckCommon)
		return m, cmd

	case showDeleteDeckMsg:
		m.list = msg.list
		m.page = newDeleteDeckPage(m.deckCommon)
		return m, cmd

	case deckCreatedMsg:
		m.list = msg.list
		m.list.InsertItem(m.list.Index(), deckItem{msg.deck})
		m.list.ResetFilter()
		m.page = newDeckBrowsePage(m.deckCommon)
		return m, nil

	case deckChangedMsg:
		m.list = msg.list
		m.list.RemoveItem(m.list.Index())
		m.list.InsertItem(m.list.Index()-1, deckItem{msg.deck})
		m.list.ResetFilter()
		m.page = newDeckBrowsePage(m.deckCommon)
		return m, nil

	case deckDeletedMsg:
		m.list = msg.list
		m.list.RemoveItem(m.list.Index())
		m.list.ResetFilter()
		m.page = newDeckBrowsePage(m.deckCommon)
		return m, nil
	}

	m.page, cmd = m.page.Update(msg)
	return m, cmd
}

func (m deckPage) View() string {
	return m.page.View()
}
