package terminal

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/eliostvs/lembrol/internal/clock"
	"github.com/eliostvs/lembrol/internal/flashcard"
)

// MODEL

type cardStatus int

const (
	cardBrowsing cardStatus = iota
	cardCreating
	cardEditing
	cardDeleting

	firstCard = 0
)

type cardItem struct {
	flashcard.Card
	// is used in the render phase to check if the card is due
	// need to be here because the default delegate don't send parameter to the description method
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

	return fmt.Sprintf("Last review %s%s", naturalTime(c.ReviewedAt), due)
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
		stats: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "stats"),
			key.WithDisabled(),
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
	stats   key.Binding
}

func newCardsModel(msg setCardsPageMsg, c clock.Clock, repo *flashcard.Repository, width, height int) cardsModel {
	keys := newCardKeys()
	delegate := list.NewDefaultDelegate()
	listModel := list.New(newCardItems(msg.deck.List(), c), &delegate, width, height)
	// force initial help width
	listModel.Help.Width = width
	listModel.Select(msg.cardIndex)
	listModel.Title = msg.deck.Name
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
			keys.stats,
			keys.study,
		}
	}

	return cardsModel{
		list:       listModel,
		clock:      c,
		deck:       msg.deck,
		repository: repo,
		keys:       keys,
		delegate:   &delegate,
		width:      width,
		height:     height,
	}
}

type cardsModel struct {
	clock      clock.Clock
	deck       flashcard.Deck
	form       form
	keys       *cardKeys
	list       list.Model
	repository *flashcard.Repository
	status     cardStatus
	delegate   *list.DefaultDelegate
	width      int
	height     int
}

// MESSAGE

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

// UPDATE

func (m cardsModel) Init() tea.Cmd {
	return func() tea.Msg {
		return initCardMsg{}
	}
}

// nolint:cyclop,gocyclo
func (m cardsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	currentCard := toCard(m.list)
	hasCards := len(m.list.Items()) != 0

	resetControls := func() {
		m.delegate.Styles.SelectedTitle = selectedTitleStyle
		m.delegate.Styles.SelectedDesc = selectedDescStyle
		m.keys.add.SetEnabled(true)
		m.keys.delete.SetEnabled(hasCards)
		m.keys.edit.SetEnabled(hasCards)
		m.keys.stats.SetEnabled(hasCards)
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
	case innerWindowSizeMsg:
		log.Printf("card.update.innerWindowsSizeMsg width=%d, height=%d\n", m.width, m.height)
		m.width, m.height = msg.Width, msg.Height
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height)
		return m, nil

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
			m.form, cmd = createCardForm("", "", m.width)
			return m, cmd

		case m.status == cardBrowsing && key.Matches(msg, m.keys.edit) && hasCards:
			m.status = cardEditing
			m.form, cmd = createCardForm(currentCard.Question, currentCard.Answer, m.width)
			return m, cmd

		case m.status == cardBrowsing && key.Matches(msg, m.keys.study) && m.deck.HasDueCards():
			return m, startReview(m.deck)

		case m.status == cardBrowsing && key.Matches(msg, m.keys.stats):
			return m, showStats(m.deck, currentCard, m.list.Index())

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
			return m, m.Init()

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

func createCardForm(question, answer string, width int) (form, tea.Cmd) {
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

	return newForm(
		newField(
			"question",
			questionInput,
			withMultiline(),
			withLabel("Front"),
		),
		newField(
			"answer",
			answerInput,
			withMultiline(),
			withLabel("Back"),
		),
	), cmd
}

func createCard(question, answer string, deck flashcard.Deck, repository *flashcard.Repository, clock clock.Clock) tea.Cmd {
	return func() tea.Msg {
		deck, card := deck.Add(question, answer)

		if err := repository.Deck.Save(deck); err != nil {
			return fail(err)
		}

		return createdCardMsg{index: 0, item: cardItem{Card: card, clock: clock}}
	}
}

func updateCard(index int, card flashcard.Card, deck flashcard.Deck, repository *flashcard.Repository, clock clock.Clock) tea.Cmd {
	return func() tea.Msg {
		if err := repository.Deck.Save(deck.Change(card)); err != nil {
			return fail(err)
		}

		return editedCardMsg{index: index, deck: deck, item: cardItem{Card: card, clock: clock}}
	}
}

func deleteCard(index int, card flashcard.Card, deck flashcard.Deck, repository *flashcard.Repository) tea.Cmd {
	return func() tea.Msg {
		if _, err := deck.Remove(card); err != nil {
			return fail(err)
		}

		if err := repository.Deck.Save(deck); err != nil {
			return fail(err)
		}

		return deletedCardMsg{index}
	}
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
