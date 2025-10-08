package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/eliostvs/lembrol/internal/clock"
	"github.com/eliostvs/lembrol/internal/flashcard"
)

// Messages

type (
	showBrowseCardMsg struct {
		list list.Model
	}

	showAddCardMsg struct {
		list list.Model
	}

	showEditCardMsg struct {
		list list.Model
		card flashcard.Card
	}

	showDeleteCardMsg struct {
		list list.Model
	}

	cardCreatedMsg struct {
		list list.Model
		card flashcard.Card
		deck flashcard.Deck
	}

	cardDeletedMsg struct {
		list list.Model
		deck flashcard.Deck
	}

	cardChangedMsg struct {
		list list.Model
		card flashcard.Card
		deck flashcard.Deck
	}
)

func showBrowseCard(model list.Model) tea.Cmd {
	return func() tea.Msg {
		return showBrowseCardMsg{list: model}
	}
}

func showAddCard(model list.Model) tea.Cmd {
	return func() tea.Msg {
		return showAddCardMsg{list: model}
	}
}

func showEditCard(model list.Model, card flashcard.Card) tea.Cmd {
	return func() tea.Msg {
		return showEditCardMsg{list: model, card: card}
	}
}

func showDeleteCard(model list.Model) tea.Cmd {
	return func() tea.Msg {
		return showDeleteCardMsg{list: model}
	}
}

func createCard(question, answer string, shared cardShared) tea.Cmd {
	return func() tea.Msg {
		deck, card := shared.deck.Add(question, answer)
		if err := shared.repository.Save(deck); err != nil {
			return fail(err)
		}
		return cardCreatedMsg{list: shared.list, card: card, deck: deck}
	}
}

func updateCard(card flashcard.Card, shared cardShared) tea.Cmd {
	return func() tea.Msg {
		deck := shared.deck.Change(card)
		if err := shared.repository.Save(deck); err != nil {
			return fail(err)
		}
		return cardChangedMsg{list: shared.list, deck: deck, card: card}
	}
}

func deleteCard(model list.Model, card flashcard.Card, shared cardShared) tea.Cmd {
	return func() tea.Msg {
		deck := shared.deck.Remove(card)
		if err := shared.repository.Save(deck); err != nil {
			return fail(err)
		}

		return cardDeletedMsg{list: model, deck: deck}
	}
}

func hasCards(m list.Model) bool {
	return len(m.Items()) != 0
}

func currentCard(m list.Model) flashcard.Card {
	item, ok := m.SelectedItem().(cardItem)
	if ok {
		return item.Card
	}

	return flashcard.Card{}
}

// Card Item

type cardItem struct {
	flashcard.Card
	// is used in the render phase to check if the card is due
	// need to be here because the list.NewDefaultDelegate don't send parameter to the description method
	clock clock.Clock
}

func (c cardItem) Title() string {
	return c.Question
}

func (c cardItem) Description() string {
	var due string

	if c.IsDue(c.clock.Now()) {
		due += " • due"
	}

	return fmt.Sprintf("Last review %s%s", naturalTime(c.LastReview), due)
}

func (c cardItem) FilterValue() string {
	return c.Question
}

func newCardItems(cards []flashcard.Card, clock clock.Clock) []list.Item {
	items := make([]list.Item, 0, len(cards))
	for _, card := range cards {
		items = append(items, cardItem{Card: card, clock: clock})
	}
	return items
}

// Browse Card

type cardBrowseKeyMap struct {
	add    key.Binding
	stats  key.Binding
	study  key.Binding
	edit   key.Binding
	delete key.Binding
}

func (k cardBrowseKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.add,
		k.study,
	}
}

func (k cardBrowseKeyMap) FullHelp() []key.Binding {
	return []key.Binding{
		k.add,
		k.edit,
		k.delete,
		k.stats,
		k.study,
	}
}

func newCardBrowsePage(shared cardShared) cardBrowsePage {
	shared.list.SetSize(shared.width-shared.styles.List.GetHorizontalFrameSize(), shared.height-shared.styles.List.GetVerticalFrameSize())
	shared.delegate.Styles.SelectedTitle = shared.styles.SelectedTitle
	shared.delegate.Styles.SelectedDesc = shared.styles.SelectedDesc

	return cardBrowsePage{
		cardShared: shared,
		keyMap: cardBrowseKeyMap{
			add: key.NewBinding(
				key.WithKeys("a"),
				key.WithHelp("a", "add"),
			),
			stats: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "stats"),
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

type cardBrowsePage struct {
	cardShared
	keyMap cardBrowseKeyMap
}

func (m cardBrowsePage) Init() tea.Cmd {
	m.Log("card-browse: init")
	return nil
}

//nolint:cyclop
func (m cardBrowsePage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.Log("cardBrowse update: %T", msg)

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height-m.styles.List.GetVerticalPadding())
		return m, nil

	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.keyMap.add):
			return m, showAddCard(m.list)

		case key.Matches(msg, m.keyMap.edit):
			return m, showEditCard(m.list, currentCard(m.list))

		case key.Matches(msg, m.keyMap.study):
			return m, startReview(m.deck)

		case key.Matches(msg, m.keyMap.delete):
			return m, showDeleteCard(m.list)

		case key.Matches(msg, m.keyMap.stats):
			return m, showStats(m.list.Index(), currentCard(m.list), m.deck)

		case key.Matches(msg, m.list.KeyMap.Quit) && m.list.FilterState() != list.FilterApplied:
			return m, showDecks(0)
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m.checkKeyMap(), cmd
}

func (m cardBrowsePage) View() string {
	m.Log("cardBrowse view: width=%d height=%d", m.width, m.height)

	return m.styles.List.Render(m.list.View())
}

func (m cardBrowsePage) checkKeyMap() cardBrowsePage {
	hasCards := hasCards(m.list)
	m.keyMap.add.SetEnabled(m.list.FilterState() == list.Unfiltered)
	m.keyMap.delete.SetEnabled(hasCards)
	m.keyMap.edit.SetEnabled(hasCards)
	m.keyMap.stats.SetEnabled(hasCards)
	m.keyMap.study.SetEnabled(m.deck.HasDueCards())
	m.list.NewStatusMessage("")
	m.list.SetFilteringEnabled(hasCards)
	m.list.SetShowStatusBar(hasCards)
	m.list.KeyMap.CursorDown.SetEnabled(hasCards)
	m.list.KeyMap.CursorDown.SetEnabled(hasCards)
	m.list.KeyMap.CursorUp.SetEnabled(hasCards)
	m.list.KeyMap.CursorUp.SetEnabled(hasCards)
	m.list.KeyMap.Filter.SetEnabled(hasCards)
	m.list.KeyMap.GoToEnd.SetEnabled(hasCards)
	m.list.KeyMap.GoToStart.SetEnabled(hasCards)
	m.list.KeyMap.NextPage.SetEnabled(hasCards)
	m.list.KeyMap.PrevPage.SetEnabled(hasCards)
	m.list.KeyMap.CloseFullHelp.SetEnabled(true)
	m.list.KeyMap.ShowFullHelp.SetEnabled(true)
	m.list.AdditionalShortHelpKeys = m.keyMap.ShortHelp
	m.list.AdditionalFullHelpKeys = m.keyMap.FullHelp

	return m
}

// Form Card

type cardFormKeyMap struct {
	submit   key.Binding
	cancel   key.Binding
	previous key.Binding
	next     key.Binding
}

func (k cardFormKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.next,
		k.previous,
		k.submit,
		k.cancel,
	}
}

func (k cardFormKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
	}
}

type field struct {
	textarea.Model
	name string
}

func (f field) IsValid() bool {
	return 0 < len(strings.TrimSpace(strings.ReplaceAll(f.Value(), "\n", "")))
}

func (f field) Update(msg tea.Msg) (field, tea.Cmd) {
	model, cmd := f.Model.Update(msg)
	f.Model = model
	return f, cmd
}

func (f field) View() string {
	return fieldStyle.Render(f.Model.View())
}

func newCardForm(question, answer string, shared Shared) cardForm {
	keyMap := cardFormKeyMap{
		submit: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "confirm"),
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
	}

	questionInput := textarea.New()
	questionInput.SetWidth(shared.width)
	questionInput.SetValue(breakLines(question))
	questionInput.Placeholder = "Enter a question"
	questionInput.ShowLineNumbers = false
	questionInput.CursorEnd()
	questionInput.Focus()

	answerInput := textarea.New()
	answerInput.SetWidth(shared.width)
	answerInput.SetValue(breakLines(answer))
	answerInput.Placeholder = "Enter an answer"
	answerInput.ShowLineNumbers = false
	answerInput.CursorEnd()
	answerInput.Blur()

	model := cardForm{
		Shared: shared,
		cursor: newCursor(1),
		fields: []field{
			{Model: questionInput, name: "question"},
			{Model: answerInput, name: "answer"},
		},
		keyMap: keyMap,
	}

	return model
}

type cardForm struct {
	Shared
	cursor cursor
	fields []field
	keyMap cardFormKeyMap
}

func (m cardForm) Init() tea.Cmd {
	return m.fields[0].Focus()
}

func (m cardForm) focus(index int) (cardForm, tea.Cmd) {
	var cmd tea.Cmd

	for i, field := range m.fields {
		if i == index {
			cmd = field.Focus()
		} else {
			field.Blur()
		}
		m.fields[i] = field
	}

	return m, cmd
}

func (m cardForm) isValid() bool {
	for _, field := range m.fields {
		if !field.IsValid() {
			return false
		}
	}
	return true
}

func (m cardForm) Value(name string) string {
	for _, field := range m.fields {
		if strings.EqualFold(field.name, name) {
			return field.Value()
		}
	}
	return ""
}

func (m cardForm) prev() (cardForm, tea.Cmd) {
	m.cursor.Up()
	return m.focus(m.cursor.Value())
}

func (m cardForm) next() (cardForm, tea.Cmd) {
	m.cursor.Down()
	return m.focus(m.cursor.Value())
}

// Update the cardForm fields inner state.
func (m cardForm) Update(msg tea.Msg) (cardForm, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.previous):
			return m.prev()

		case key.Matches(msg, m.keyMap.next):
			return m.next()

		case key.Matches(msg, m.keyMap.cancel):
			return m, cancelForm()

		case key.Matches(msg, m.keyMap.submit):
			if m.isValid() {
				return m, submitForm(m)
			}
		}
	}

	return m.updateFields(msg)
}

func (m cardForm) updateFields(msg tea.Msg) (cardForm, tea.Cmd) {
	var cmd tea.Cmd

	for i, field := range m.fields {
		if field.Focused() {
			m.fields[i], cmd = field.Update(msg)
		}
	}

	return m, cmd
}

func (m cardForm) View() string {
	footer := lipgloss.
		NewStyle().
		Width(m.width).
		Margin(1, 2).
		Render(renderHelp(m.keyMap, m.width, false))

	input := lipgloss.NewStyle().
		Height(m.height - lipgloss.Height(footer)).
		Width(m.width).
		Render(m.fieldsView())

	return lipgloss.JoinVertical(lipgloss.Top, input, footer)
}

func (m cardForm) fieldsView() string {
	content := make([]string, len(m.fields))

	for i, field := range m.fields {
		content[i] = field.View()
	}

	return lipgloss.JoinVertical(lipgloss.Top, content...)
}

// Add Card

func newCardAddPage(shared cardShared) cardAddPage {
	return cardAddPage{form: newCardForm("", "", shared.Shared), cardShared: shared}
}

type cardAddPage struct {
	cardShared
	form cardForm
}

func (m cardAddPage) Init() tea.Cmd {
	m.Log("card-add: init")

	return m.form.Init()
}

func (m cardAddPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.Log("cardAdd: %T", msg)

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case submittedFormMsg[cardForm]:
		return m, tea.Batch(
			showLoading(m.deck.Name, "Creating card..."),
			createCard(msg.data.Value("question"), m.form.Value("answer"), m.cardShared),
		)

	case canceledFormMsg:
		return m, showBrowseCard(m.list)
	}

	m.form, cmd = m.form.Update(msg)
	return m, cmd
}

func (m cardAddPage) View() string {
	m.Log("cardAdd view: width=%d height=%d", m.width, m.height)

	header := m.styles.Title.
		Margin(1, 0, 0, 2).
		Render(m.deck.Name)

	subTitle := m.styles.DimmedTitle.
		Margin(1, 0, 0, 2).
		Render("Add")

	m.form.height = m.height - lipgloss.Height(header) - lipgloss.Height(subTitle)
	form := m.styles.Text.Render(m.form.View())

	return lipgloss.JoinVertical(lipgloss.Top, header, subTitle, form)
}

// Edit Card

func newCardEditPage(card flashcard.Card, shared cardShared) cardEditPage {
	return cardEditPage{card: card, form: newCardForm(card.Question, card.Answer, shared.Shared), cardShared: shared}
}

type cardEditPage struct {
	cardShared
	card flashcard.Card
	form cardForm
}

func (m cardEditPage) Init() tea.Cmd {
	m.Log("card-edit: init")
	return nil
}

func (m cardEditPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.Log("cardEdit update: %T", msg)

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case submittedFormMsg[cardForm]:
		m.card.Answer = msg.data.Value("answer")
		m.card.Question = msg.data.Value("question")

		return m, tea.Batch(
			showLoading(m.deck.Name, "Updating card..."),
			updateCard(m.card, m.cardShared),
		)

	case canceledFormMsg:
		return m, showBrowseCard(m.list)
	}

	m.form, cmd = m.form.Update(msg)
	return m, cmd
}

func (m cardEditPage) View() string {
	m.Log("cardEdit view: width=%d height=%d", m.width, m.height)

	header := m.styles.Title.
		Margin(1, 0, 0, 2).
		Render(m.deck.Name)

	subTitle := m.styles.DimmedTitle.
		Margin(1, 0, 0, 2).
		Render("Edit")

	m.form.height = m.height - lipgloss.Height(header) - lipgloss.Height(subTitle)
	form := m.styles.Text.Render(m.form.View())

	return lipgloss.JoinVertical(lipgloss.Top, header, subTitle, form)
}

// Delete Card

type cardDeleteKeyMap struct {
	confirm key.Binding
	cancel  key.Binding
}

func (k cardDeleteKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.confirm,
		k.cancel,
	}
}

func (k cardDeleteKeyMap) FullHelp() []key.Binding {
	return []key.Binding{
		k.confirm,
		k.cancel,
	}
}

func newDeleteCardPage(shared cardShared) cardDeletePage {
	keyMap := cardDeleteKeyMap{
		cancel: key.NewBinding(
			key.WithKeys(tea.KeyEsc.String()),
			key.WithHelp("esc", "cancel"),
		),
		confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
	}
	shared.delegate.Styles.SelectedTitle = shared.styles.DeletedTitle
	shared.delegate.Styles.SelectedDesc = shared.styles.DeletedDesc

	shared.list.SetSize(shared.width-shared.styles.List.GetHorizontalFrameSize(), shared.height-shared.styles.List.GetVerticalFrameSize())
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
	shared.list.NewStatusMessage(shared.styles.DeletedStatus.Render("Delete this card?"))

	return cardDeletePage{
		cardShared: shared,
		keyMap:     keyMap,
	}
}

type cardDeletePage struct {
	cardShared
	keyMap cardDeleteKeyMap
}

func (m cardDeletePage) Init() tea.Cmd {
	m.Log("card-delete: init")
	return nil
}

func (m cardDeletePage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.Log("cardDelete update: %T", msg)

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.confirm):
			return m, tea.Batch(
				showLoading("Cards", "Deleting card..."),
				deleteCard(m.list, currentCard(m.list), m.cardShared),
			)

		case key.Matches(msg, m.keyMap.cancel):
			return m, showBrowseCard(m.list)
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m cardDeletePage) View() string {
	m.Log("cardDelete view: width=%d height=%d", m.width, m.height)

	return m.styles.List.Render(m.list.View())
}

// Card Page

func newCardPage(parent Shared, deck flashcard.Deck) cardPage {
	delegate := list.NewDefaultDelegate()
	shared := cardShared{
		Shared:   parent,
		delegate: &delegate,
		list:     list.New(newCardItems(deck.List(), parent.clock), &delegate, parent.width, parent.height),
		deck:     deck,
	}
	shared.list.SetSize(shared.width, shared.height)
	shared.list.Title = deck.Name
	shared.list.Styles.NoItems = shared.list.Styles.NoItems.Copy().Margin(0, 2)

	return cardPage{
		cardShared: shared,
		page:       newCardBrowsePage(shared),
	}
}

type cardShared struct {
	Shared
	deck     flashcard.Deck
	delegate *list.DefaultDelegate
	list     list.Model
}

type cardPage struct {
	cardShared
	page tea.Model
}

func (m cardPage) Init() tea.Cmd {
	m.Log("card: init")

	return m.page.Init()
}

func (m cardPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.Log("card update: %T", msg)

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case showLoadingMsg:
		m.page = newLoadingPage(m.Shared, msg.title, msg.description)
		return m, m.page.Init()

	case showBrowseCardMsg:
		m.list = msg.list
		m.page = newCardBrowsePage(m.cardShared)
		return m, nil

	case showAddCardMsg:
		m.list = msg.list
		m.page = newCardAddPage(m.cardShared)
		return m, cmd

	case showEditCardMsg:
		m.list = msg.list
		m.page = newCardEditPage(msg.card, m.cardShared)
		return m, cmd

	case showDeleteCardMsg:
		m.list = msg.list
		m.page = newDeleteCardPage(m.cardShared)
		return m, cmd

	case cardCreatedMsg:
		m.list = msg.list
		m.deck = msg.deck
		m.list.InsertItem(m.list.Index(), cardItem{Card: msg.card, clock: m.clock})
		m.list.ResetFilter()
		m.page = newCardBrowsePage(m.cardShared)
		return m, nil

	case cardChangedMsg:
		m.list = msg.list
		m.deck = msg.deck
		m.list.RemoveItem(m.list.Index())
		m.list.InsertItem(m.list.Index()-1, cardItem{Card: msg.card, clock: m.clock})
		m.list.ResetFilter()
		m.page = newCardBrowsePage(m.cardShared)
		return m, nil

	case cardDeletedMsg:
		m.list = msg.list
		m.deck = msg.deck
		m.list.RemoveItem(m.list.Index())
		m.list.ResetFilter()
		m.page = newCardBrowsePage(m.cardShared)
		return m, nil
	}

	m.page, cmd = m.page.Update(msg)
	return m, cmd
}

func (m cardPage) View() string {
	m.Log("card view: width=%d height=%d", m.width, m.height)

	return m.page.View()
}
