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

type cardStatus int

const (
	cardBrowsing cardStatus = iota
	cardCreating
	cardEditing
	cardDeleting
)

// ITEM

type cardItem struct {
	flashcard.Card
	clock flashcard.Clock
}

func (c cardItem) Title() string {
	return c.Question
}

func (c cardItem) Description() string {
	var due string
	if c.Due(c.clock.Now()) {
		due += " • due"
	}

	return fmt.Sprintf("Last review %s%s", naturalTime(c.ReviewedAt), due)
}

func (c cardItem) FilterValue() string {
	return c.Question
}

func newCardItems(cards []flashcard.Card, clock flashcard.Clock) []list.Item {
	items := make([]list.Item, 0, len(cards))
	for _, card := range cards {
		items = append(items, cardItem{Card: card, clock: clock})
	}
	return items
}

// KEYS

func newCardKeys() *cardKeys {
	return &cardKeys{
		add: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add"),
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
		confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
			key.WithDisabled(),
		),
	}
}

type cardKeys struct {
	add     key.Binding
	study   key.Binding
	edit    key.Binding
	delete  key.Binding
	confirm key.Binding
}

// MODEL

func newCardsModel(deck flashcard.Deck, clock flashcard.Clock, repository *flashcard.DeckRepository, v viewport) cardsModel {
	keys := newCardKeys()
	delegate := list.NewDefaultDelegate()
	listModel := list.New(newCardItems(deck.List(), clock), &delegate, 0, 0)
	listModel.Title = deck.Name
	listModel.Styles.Title = titleStyle
	listModel.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			keys.add,
			keys.study,
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
	listModel.DisableQuitKeybindings()

	return cardsModel{
		list:       listModel,
		clock:      clock,
		deck:       deck,
		repository: repository,
		keys:       keys,
		delegate:   &delegate,
		viewport:   v,
	}
}

type cardsModel struct {
	clock      flashcard.Clock
	deck       flashcard.Deck
	form       Form
	keys       *cardKeys
	list       list.Model
	repository *flashcard.DeckRepository
	status     cardStatus
	delegate   *list.DefaultDelegate
	viewport   viewport
}

// VIEW

func (m cardsModel) View() string {
	switch m.status {
	case cardCreating:
		content := titleStyle.Render(m.deck.Name)
		content += m.form.view()
		return largePaddingStyle.Render(content)

	case cardEditing:
		content := titleStyle.Render(m.deck.Name)
		content += m.form.view()
		return largePaddingStyle.Render(content)

	case cardBrowsing, cardDeleting:
		fallthrough

	default:
		return midPaddingStyle.Render(m.list.View())
	}
}

// INIT

func (m cardsModel) init() tea.Cmd {
	return func() tea.Msg {
		return initCardMsg{}
	}
}

// UPDATE

type (
	initCardMsg struct{}

	createdCardMsg struct {
		index int
		item  cardItem
	}

	deletedCardMsg struct {
		index int
	}

	editedCardMsg struct {
		index int
		item  cardItem
		deck  flashcard.Deck
	}
)

// nolint:cyclop,gocyclo
func (m cardsModel) Update(msg tea.Msg) (cardsModel, tea.Cmd) {
	var cmd tea.Cmd

	m.list.SetWidth(m.viewport.width)
	m.list.SetHeight(m.viewport.height)

	currentCard := toCard(m.list)
	hasCards := len(m.list.Items()) != 0

	resetControls := func() {
		m.delegate.Styles.SelectedTitle = selectedTitleStyle
		m.delegate.Styles.SelectedDesc = selectedDescStyle
		m.keys.add.SetEnabled(true)
		m.keys.delete.SetEnabled(hasCards)
		m.keys.edit.SetEnabled(hasCards)
		m.keys.study.SetEnabled(m.deck.HasDueCards())
		m.keys.confirm.SetEnabled(m.status == cardDeleting)
		m.list.NewStatusMessage("")
		m.list.SetFilteringEnabled(hasCards)
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
	}

	switch msg := msg.(type) {
	case initCardMsg, canceledFormMsg:
		m.status = cardBrowsing
		resetControls()
		return m, nil

	case createdCardMsg:
		m.status = cardBrowsing
		m.list.InsertItem(msg.index, msg.item)
		resetControls()
		return m, nil

	case editedCardMsg:
		m.status = cardBrowsing
		m.list.RemoveItem(msg.index)
		m.list.InsertItem(msg.index-1, msg.item)
		resetControls()
		return m, nil

	case deletedCardMsg:
		m.status = cardBrowsing
		m.list.RemoveItem(msg.index)
		resetControls()
		return m, nil

	case submittedFormMsg:
		if m.status == cardEditing {
			currentCard.Answer = m.form.Value("answer")
			currentCard.Question = m.form.Value("question")
			return m, updateCard(m.list.Index(), currentCard, m.deck, m.repository, m.clock)
		}

		if m.status == cardCreating {
			return m, createCard(m.form.Value("question"), m.form.Value("answer"), m.deck, m.repository, m.clock)
		}

	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case m.status == cardBrowsing && key.Matches(msg, m.keys.add):
			m.status = cardCreating
			m.form, cmd = createCardForm("", "", m.viewport.width)
			return m, cmd

		case m.status == cardBrowsing && key.Matches(msg, m.keys.edit) && hasCards:
			m.status = cardEditing
			m.form, cmd = createCardForm(currentCard.Question, currentCard.Answer, m.viewport.width)
			return m, cmd

		case m.status == cardBrowsing && key.Matches(msg, m.keys.study) && m.deck.HasDueCards():
			return m, startReview(m.deck)

		case m.status == cardBrowsing && key.Matches(msg, m.list.KeyMap.Quit) && m.list.FilterState() != list.FilterApplied:
			return m, showDecks

		case m.status == cardBrowsing && key.Matches(msg, m.keys.delete) && hasCards:
			m.status = cardDeleting
			m.delegate.Styles.SelectedTitle = deletedTitle
			m.delegate.Styles.SelectedDesc = deletedDesc
			m.keys.add.SetEnabled(false)
			m.keys.delete.SetEnabled(false)
			m.keys.edit.SetEnabled(false)
			m.keys.study.SetEnabled(false)
			m.keys.confirm.SetEnabled(true)
			m.list.NewStatusMessage(Red.Render("Delete this card?"))
			m.list.KeyMap.CursorDown.SetEnabled(false)
			m.list.KeyMap.CursorUp.SetEnabled(false)
			m.list.KeyMap.Filter.SetEnabled(false)
			m.list.KeyMap.GoToEnd.SetEnabled(false)
			m.list.KeyMap.GoToStart.SetEnabled(false)
			m.list.KeyMap.NextPage.SetEnabled(false)
			m.list.KeyMap.PrevPage.SetEnabled(false)
			m.list.KeyMap.CloseFullHelp.SetEnabled(false)
			m.list.KeyMap.ShowFullHelp.SetEnabled(false)
			return m, nil

		case m.status == cardDeleting && key.Matches(msg, m.list.KeyMap.Quit):
			return m, m.init()

		case m.status == cardDeleting && key.Matches(msg, m.keys.confirm):
			return m, deleteCard(m.list.Index(), currentCard, m.deck, m.repository)

			// the only two actions in delete state should confirm or cancel
		case m.status == cardDeleting:
			return m, nil
		}
	}

	if m.status == cardEditing || m.status == cardCreating {
		m.form, cmd = m.form.Update(msg)
		return m, cmd
	}

	m.list, cmd = m.list.Update(msg)
	resetControls()
	return m, cmd
}

func toCard(l list.Model) flashcard.Card {
	item, ok := l.SelectedItem().(cardItem)
	if ok {
		return item.Card
	}
	return flashcard.Card{}
}

func createCardForm(question, answer string, width int) (Form, tea.Cmd) {
	var cmd tea.Cmd

	questionInput := textinput.New()
	questionInput.SetValue(question)
	questionInput.Placeholder = "Enter a question"
	questionInput.PromptStyle = Fuchsia
	questionInput.TextStyle = Fuchsia
	questionInput.Width = width
	questionInput.CursorEnd()
	cmd = questionInput.Focus()

	answerInput := textinput.New()
	answerInput.SetValue(answer)
	answerInput.Placeholder = "Enter an answer"
	answerInput.PromptStyle = DarkFuchsia
	answerInput.TextStyle = DarkFuchsia
	answerInput.Width = width
	answerInput.CursorEnd()
	answerInput.Blur()

	return NewForm(
		NewField(
			"question",
			questionInput,
			WithMultiline(),
			WithLabel("Front"),
		),
		NewField(
			"answer",
			answerInput,
			WithMultiline(),
			WithLabel("Back"),
		),
	), cmd
}

func createCard(question, answer string, deck flashcard.Deck, repository *flashcard.DeckRepository, clock flashcard.Clock) tea.Cmd {
	return func() tea.Msg {
		deck, card := deck.Add(question, answer)

		if err := repository.Save(deck); err != nil {
			return failed(err)
		}

		return createdCardMsg{index: 0, item: cardItem{Card: card, clock: clock}}
	}
}

func updateCard(index int, card flashcard.Card, deck flashcard.Deck, repository *flashcard.DeckRepository, clock flashcard.Clock) tea.Cmd {
	return func() tea.Msg {
		if err := repository.Save(deck.Change(card)); err != nil {
			return failed(err)
		}

		return editedCardMsg{index: index, deck: deck, item: cardItem{Card: card, clock: clock}}
	}
}

func deleteCard(index int, card flashcard.Card, deck flashcard.Deck, repository *flashcard.DeckRepository) tea.Cmd {
	return func() tea.Msg {
		if _, err := deck.Remove(card); err != nil {
			return failed(err)
		}

		if err := repository.Save(deck); err != nil {
			return failed(err)
		}

		return deletedCardMsg{index}
	}
}
