package terminal

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/eliostvs/lembrol/internal/flashcard"
)

// MODEL

func newStatsModel(msg setStatsPageMsg, repo *flashcard.Repository, viewport viewport) statsModel {
	return statsModel{
		cardIndex:  msg.cardIndex,
		deck:       msg.deck,
		card:       msg.card,
		repository: repo,
		viewport:   viewport,
	}
}

type statsModel struct {
	cardIndex  int
	deck       flashcard.Deck
	card       flashcard.Card
	repository *flashcard.Repository
	viewport   viewport
}

// INIT

func (m statsModel) Init() tea.Cmd {
	return loadStats(m.repository, m.deck, m.card)
}

// UPDATE

type (
	statsLoadedMsg struct {
		stats []flashcard.Stats
	}
)

func (m statsModel) Update(msg tea.Msg) (statsModel, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		return m, showCards(m.deck, m.cardIndex)
	}

	return m, nil
}

func loadStats(repo *flashcard.Repository, deck flashcard.Deck, card flashcard.Card) tea.Cmd {
	return func() tea.Msg {
		stats, err := repo.Stats.Find(deck, card)
		if err != nil {
			return failed(err)
		}

		return statsLoadedMsg{stats: stats}
	}
}

// VIEW

func (m statsModel) View() string {
	// show loading page
	// show stats
	return "No stats"
}
