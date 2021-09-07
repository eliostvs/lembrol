package terminal

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/eliostvs/remembercli/internal/flashcard"
)

// MODEL

type deckStatus int

const (
	deckBrowsing deckStatus = iota
	deckCreating
	deckDeleting
	deckEditing
)

func (s deckStatus) template() string {
	return []string{
		"deckShow",
		"deckCreate",
		"deckDelete",
		"deckEdit",
	}[s]
}

func newDeckModel(decks []flashcard.Deck, r *flashcard.Repository) deckModel {
	return deckModel{
		Decks:      decks,
		Page:       newPosition(len(decks)),
		repository: r,
	}
}

type deckModel struct {
	Decks []flashcard.Deck
	Form  Form
	Page  position

	status     deckStatus
	repository *flashcard.Repository
}

// VIEW

func (m deckModel) Template() string {
	return m.status.template()
}

// UPDATE

type deletedDeckMsg struct {
	flashcard.Deck
}

type renamedDeckMsg struct {
	flashcard.Deck
}

type createdDeckMsg struct {
	flashcard.Deck
}

// nolint:cyclop,gocognit
func (m deckModel) Update(msg tea.Msg) (deckModel, tea.Cmd) {
	hasDecks := hasDecks(m.Decks)
	currentDeck := currentDeck(m.Page.Item(), m.Decks)

	switch msg := msg.(type) {
	case createdDeckMsg:
		m.status = deckBrowsing
		m.Decks = append(m.Decks, msg.Deck)
		m.Page = m.Page.Increase()
		return m, nil

	case renamedDeckMsg:
		m.status = deckBrowsing
		m.Decks = updateDeck(m.Decks, msg.Deck)
		return m, nil

	case deletedDeckMsg:
		m.status = deckBrowsing
		m.Decks = removeDeck(m.Decks, msg.Deck)
		m.Page = m.Page.Decrease()
		return m, nil

	case submittedFormMsg:
		if m.status == deckEditing {
			return m, renameDeck(m.Form.Value("name"), currentDeck, m.repository)
		}

		if m.status == deckCreating {
			return m, createDeck(m.Form.Value("name"), m.repository)
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "a":
			if m.status == deckBrowsing {
				m.status = deckCreating
				m.Form = createDeckForm("", "> ")
				return m, textinput.Blink
			}

		case "q":
			if m.status == deckBrowsing || m.status == deckDeleting {
				return m, exit
			}

		case "r":
			if m.status == deckBrowsing && hasDecks {
				m.status = deckEditing
				m.Form = createDeckForm(currentDeck.Name, " ")
				return m, textinput.Blink
			}

		case "s":
			if m.status == deckBrowsing && currentDeck.HasDueCards() {
				return m, startReview(currentDeck)
			}

		case "x":
			if m.status == deckBrowsing && hasDecks {
				m.status = deckDeleting
				return m, nil
			}

		case tea.KeyUp.String(), "k", tea.KeyDown.String(), "j", tea.KeyLeft.String(), tea.KeyRight.String(), "l", "h":
			if m.status == deckBrowsing {
				m.Page = m.Page.Update(msg)
				return m, nil
			}

		case tea.KeyEsc.String():
			if m.status != deckBrowsing {
				m.status = deckBrowsing
				return m, nil
			}

		case tea.KeyEnter.String():
			if m.status == deckDeleting {
				return m, deleteDeck(currentDeck, m.repository)
			}

			if m.status == deckBrowsing && hasDecks {
				return m, showCards(currentDeck)
			}
		}
	}

	var cmd tea.Cmd
	m.Form, cmd = m.Form.Update(msg)
	return m, cmd
}

func hasDecks(decks []flashcard.Deck) bool {
	return len(decks) > 0
}

func currentDeck(index int, decks []flashcard.Deck) flashcard.Deck {
	for i, deck := range decks {
		if index == i {
			return deck
		}
	}
	return flashcard.Deck{}
}

func updateDeck(old []flashcard.Deck, changed flashcard.Deck) (decks []flashcard.Deck) {
	for _, deck := range old {
		if deck.Id() == changed.Id() {
			decks = append(decks, changed)
		} else {
			decks = append(decks, deck)
		}
	}
	return decks
}

func removeDeck(original []flashcard.Deck, deleted flashcard.Deck) (decks []flashcard.Deck) {
	for _, deck := range original {
		if deck.Id() != deleted.Id() {
			decks = append(decks, deck)
		}
	}
	return decks
}

func createDeckForm(name, prompt string) Form {
	input := textinput.NewModel()
	input.CharLimit = 30
	input.SetValue(name)
	input.CursorEnd()
	input.Prompt = prompt
	input.TextStyle = DarkGreen
	input.PromptStyle = DarkGreen
	input.Focus()
	return NewForm(NewField("name", input))
}

func createDeck(name string, repo *flashcard.Repository) tea.Cmd {
	return func() tea.Msg {
		deck, err := repo.Create(name)
		if err != nil {
			return failed(err)
		}
		return createdDeckMsg{deck}
	}
}

func renameDeck(name string, deck flashcard.Deck, repo *flashcard.Repository) tea.Cmd {
	return func() tea.Msg {
		deck.Name = name
		if err := repo.Save(deck); err != nil {
			return failed(err)
		}
		return renamedDeckMsg{deck}
	}
}

func deleteDeck(deck flashcard.Deck, repo *flashcard.Repository) tea.Cmd {
	return func() tea.Msg {
		if err := repo.Remove(deck); err != nil {
			return failed(err)
		}
		return deletedDeckMsg{deck}
	}
}
