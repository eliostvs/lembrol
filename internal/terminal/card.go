package terminal

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/eliostvs/remembercli/internal/flashcard"
)

// MODEL

type cardStatus int

const (
	cardBrowsing cardStatus = iota
	cardCreating
	cardEditing
	cardDeleting
)

func (s cardStatus) template() string {
	return []string{
		"cardShow",
		"cardForm",
		"cardForm",
		"cardDelete",
	}[s]
}

func newCardModel(deck flashcard.Deck, clock flashcard.Clock, repository *flashcard.Repository) cardModel {
	return cardModel{
		Cards:      deck.List(),
		Clock:      clock,
		Deck:       deck,
		Page:       newPosition(deck.Total()),
		repository: repository,
	}
}

type cardModel struct {
	Cards []flashcard.Card
	Clock flashcard.Clock
	Deck  flashcard.Deck
	Form  form
	Page  position
	Title string

	repository *flashcard.Repository
	status     cardStatus
}

// VIEW

func (m cardModel) Template() string {
	return m.status.template()
}

// UPDATE

type (
	createdCardMsg struct {
		flashcard.Card
	}

	deletedCardMsg struct {
		flashcard.Card
	}

	editedCardMsg struct {
		flashcard.Card
	}
)

// nolint:cyclop
func (m cardModel) Update(width int, msg tea.Msg) (cardModel, tea.Cmd) {
	var cmd tea.Cmd

	hasCards := hasCards(m.Cards)
	currentCard := currentCard(m.Page.Item(), m.Cards)

	switch msg := msg.(type) {
	case createdCardMsg:
		m.Cards = append(m.Cards, msg.Card)
		m.status = cardBrowsing
		m.Page = m.Page.Increase()
		return m, nil

	case editedCardMsg:
		m.Cards = updateCardList(m.Cards, msg.Card)
		m.status = cardBrowsing
		return m, nil

	case deletedCardMsg:
		m.Cards = removeCard(m.Cards, msg.Card)
		m.status = cardBrowsing
		m.Page = m.Page.Decrease()
		return m, nil

	case submittedFormMsg:
		if m.status == cardEditing {
			currentCard.Answer = m.Form.Value("answer")
			currentCard.Question = m.Form.Value("question")
			m.Deck = m.Deck.Change(currentCard)
			return m, updateCard(currentCard, m.Deck, m.repository)
		}

		if m.status == cardCreating {
			return m, createCard(m.Form.Value("question"), m.Form.Value("answer"), m.Deck, m.repository)
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "a":
			if m.status == cardBrowsing {
				m.Form, cmd = createCardForm("", "", width)
				m.status = cardCreating
				m.Title = "Add Card"
				return m, cmd
			}

		case "e":
			if m.status == cardBrowsing {
				m.Form, cmd = createCardForm(currentCard.Question, currentCard.Answer, width)
				m.status = cardEditing
				m.Title = "Edit Card"
				return m, cmd
			}

		case "q":
			if m.status == cardBrowsing || m.status == cardDeleting {
				return m, exit
			}

		case "s":
			if m.status == cardBrowsing && m.Deck.HasDueCards() {
				return m, startReview(m.Deck)
			}

		case "x":
			if m.status == cardBrowsing && hasCards {
				m.status = cardDeleting
				return m, nil
			}

		case tea.KeyUp.String(), "k", tea.KeyDown.String(), "j", tea.KeyLeft.String(), tea.KeyRight.String(), "l", "h":
			if m.status == cardBrowsing {
				m.Page = m.Page.Update(msg)
			}

		case tea.KeyEsc.String():
			if m.status == cardBrowsing {
				return m, showDecks
			}

			m.status = cardBrowsing
			return m, nil

		case tea.KeyEnter.String():
			if m.status == cardDeleting {
				return m, deleteCard(currentCard, m.Deck, m.repository)
			}
		}
	}

	m.Form, cmd = m.Form.Width(width).Update(msg)
	return m, cmd
}

func hasCards(cards []flashcard.Card) bool {
	return len(cards) > 0
}

func currentCard(index int, cards []flashcard.Card) flashcard.Card {
	for i, card := range cards {
		if index == i {
			return card
		}
	}
	return flashcard.Card{}
}

func updateCardList(original []flashcard.Card, changed flashcard.Card) (cards []flashcard.Card) {
	for _, card := range original {
		if card.Id() == changed.Id() {
			cards = append(cards, changed)
		} else {
			cards = append(cards, card)
		}
	}
	return cards
}

func removeCard(original []flashcard.Card, deleted flashcard.Card) (cards []flashcard.Card) {
	for _, card := range original {
		if card.Id() != deleted.Id() {
			cards = append(cards, card)
		}
	}
	return cards
}

func createCardForm(question, answer string, width int) (form, tea.Cmd) {
	var cmd tea.Cmd

	questionInput := textinput.NewModel()
	questionInput.SetValue(question)
	questionInput.Placeholder = "Enter a question"
	questionInput.PromptStyle = Fuchsia
	questionInput.TextStyle = Fuchsia
	questionInput.Width = width
	questionInput.CursorEnd()
	cmd = questionInput.Focus()

	answerInput := textinput.NewModel()
	answerInput.SetValue(answer)
	answerInput.Placeholder = "Enter an answer"
	answerInput.PromptStyle = DarkFuchsia
	answerInput.TextStyle = DarkFuchsia
	answerInput.Width = width
	answerInput.CursorEnd()
	answerInput.Blur()

	return newForm(
		newField("question", questionInput),
		newField("answer", answerInput),
	), cmd
}

func createCard(question, answer string, deck flashcard.Deck, repository *flashcard.Repository) tea.Cmd {
	return func() tea.Msg {
		deck, card := deck.Add(question, answer)

		if err := repository.Save(deck); err != nil {
			return failed(err)
		}

		return createdCardMsg{card}
	}
}

func updateCard(card flashcard.Card, deck flashcard.Deck, repository *flashcard.Repository) tea.Cmd {
	return func() tea.Msg {
		if err := repository.Save(deck); err != nil {
			return failed(err)
		}

		return editedCardMsg{card}
	}
}

func deleteCard(card flashcard.Card, deck flashcard.Deck, repository *flashcard.Repository) tea.Cmd {
	return func() tea.Msg {
		if _, err := deck.Remove(card); err != nil {
			return failed(err)
		}

		if err := repository.Save(deck); err != nil {
			return failed(err)
		}

		return deletedCardMsg{card}
	}
}
