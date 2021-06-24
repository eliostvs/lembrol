package terminal

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/eliostvs/remembercli/internal/flashcard"
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

func showCards(d *flashcard.Deck) tea.Cmd {
	return func() tea.Msg {
		return setCardsPageMsg{d}
	}
}

type setCardsPageMsg struct {
	*flashcard.Deck
}

func startReview(d *flashcard.Deck) tea.Cmd {
	return func() tea.Msg {
		return setReviewPageMsg{d}
	}
}

type setReviewPageMsg struct {
	*flashcard.Deck
}

type setQuitPageMsg struct{}

func exit() tea.Msg {
	return setQuitPageMsg{}
}
