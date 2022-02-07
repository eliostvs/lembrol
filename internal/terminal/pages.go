package terminal

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/eliostvs/lembrol/internal/flashcard"
)

type page int

const (
	Loading page = iota
	Decks
	Cards
	Review
	Quit
	Error
)

func failed(err error) tea.Msg {
	return setErrorPageMsg{Error: err.Error()}
}

type setErrorPageMsg struct {
	Error string
}

func showDecks() tea.Msg {
	return setDecksPageMsg{}
}

type setDecksPageMsg struct{}

func showCards(deck flashcard.Deck, card int) tea.Cmd {
	return func() tea.Msg {
		return setCardsPageMsg{deck: deck, card: card}
	}
}

type setCardsPageMsg struct {
	deck flashcard.Deck
	card int
}

func startReview(d flashcard.Deck) tea.Cmd {
	return func() tea.Msg {
		return setReviewPageMsg{d}
	}
}

type setReviewPageMsg struct {
	flashcard.Deck
}

type setQuitPageMsg struct{}

func exitCmd() tea.Msg {
	return setQuitPageMsg{}
}
