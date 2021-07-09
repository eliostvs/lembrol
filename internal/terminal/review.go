package terminal

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/eliostvs/remembercli/internal/flashcard"
)

// MODEL

type reviewStatus int

const (
	reviewQuestion reviewStatus = iota
	reviewAnswer
	reviewFinished
)

func (s reviewStatus) template() string {
	return []string{
		"question",
		"answer",
		"review",
	}[s]
}

func newReviewModel(review *flashcard.Review, repository *flashcard.Repository) reviewModel {
	return reviewModel{
		Review:     review,
		repository: repository,
	}
}

type reviewModel struct {
	Review *flashcard.Review

	repository *flashcard.Repository
	status     reviewStatus
}

// VIEW

func (m reviewModel) Template() string {
	return m.status.template()
}

// UPDATE

type (
	scoredMsg   struct{}
	reviewedMsg struct{}
)

// nolint:cyclop
func (m reviewModel) Update(msg tea.Msg) (reviewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case scoredMsg:
		m.status = reviewQuestion
		return m, nil

	case reviewedMsg:
		m.status = reviewFinished
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyEsc.String():
			if m.status == reviewFinished {
				return m, showDecks
			}
			return m, showCards(m.Review.Deck())

		case tea.KeyEnter.String():
			if m.status == reviewQuestion {
				m.status = reviewAnswer
				return m, nil
			}

		case "s":
			if m.status == reviewQuestion && m.Review.Total() > 1 {
				return m, skipCard(m.Review)
			}

		case "q":
			return m, exit

		case "0", "1", "2", "3", "4":
			if m.status == reviewAnswer {
				return m, scoreCard(m.repository, m.Review, msg.String())
			}
		}
	}

	return m, nil
}

func skipCard(review *flashcard.Review) tea.Cmd {
	return func() tea.Msg {
		if err := review.Skip(); err != nil {
			return failed(err)
		}
		return nil
	}
}

func scoreCard(repo *flashcard.Repository, review *flashcard.Review, input string) tea.Cmd {
	return func() tea.Msg {
		score, err := flashcard.NewReviewScore(input)
		if err != nil {
			return failed(err)
		}

		stats, err := review.Rate(score)
		if err != nil {
			return failed(err)
		}

		if err = repo.Save(review.Deck()); err != nil {
			return failed(err)
		}

		if stats != nil {
			_ = repo.SaveStats(stats)
		}

		if review.Left() == 0 {
			return reviewedMsg{}
		}

		return scoredMsg{}
	}
}
