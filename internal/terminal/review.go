package terminal

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/eliostvs/lembrol/internal/flashcard"
)

// MODEL

type reviewStatus int

const (
	reviewQuestion reviewStatus = iota
	reviewAnswer
	reviewFinished
)

func newReviewKeys() *reviewKeys {
	return &reviewKeys{
		quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q", "quit"),
		),
		score: key.NewBinding(
			key.WithKeys("0", "1", "2", "3", "4"),
			key.WithHelp("1", "score"),
		),
		answer: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "answer"),
		),
		skip: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "skip"),
			key.WithDisabled(),
		),
		again: key.NewBinding(
			key.WithHelp("0", "again"),
			key.WithDisabled(),
		),
		hard: key.NewBinding(
			key.WithHelp("1", "hard"),
			key.WithDisabled(),
		),
		normal: key.NewBinding(
			key.WithHelp("2", "normal"),
			key.WithDisabled(),
		),
		easy: key.NewBinding(
			key.WithHelp("3", "easy"),
			key.WithDisabled(),
		),
		veryEasy: key.NewBinding(
			key.WithHelp("4", "very easy"),
			key.WithDisabled(),
		),
		showFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "more"),
		),
		closeFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "close help"),
		),
	}
}

type reviewKeys struct {
	quit          key.Binding
	answer        key.Binding
	skip          key.Binding
	score         key.Binding
	again         key.Binding
	hard          key.Binding
	normal        key.Binding
	easy          key.Binding
	veryEasy      key.Binding
	showFullHelp  key.Binding
	closeFullHelp key.Binding
}

func (k *reviewKeys) ShortHelp() []key.Binding {
	return []key.Binding{
		k.skip,
		k.answer,
		k.hard,
		k.normal,
		k.easy,
		k.veryEasy,
		k.quit,
		k.showFullHelp,
	}
}

func (k *reviewKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.skip,
			k.again,
		},
		{
			k.answer,
			k.hard,
			k.normal,
			k.easy,
			k.veryEasy,
		},
		{
			k.quit,
			k.closeFullHelp,
		},
	}
}

func newReviewModel(review flashcard.Review, repository *flashcard.Repository, v viewport) reviewModel {
	helpModel := help.New()
	helpModel.Width = v.Width

	return reviewModel{
		review:     review,
		repository: repository,
		keys:       newReviewKeys(),
		help:       helpModel,
		viewport:   v,
	}
}

type reviewModel struct {
	review     flashcard.Review
	help       help.Model
	repository *flashcard.Repository
	status     reviewStatus
	keys       *reviewKeys
	viewport   viewport
}

// INIT

func (m reviewModel) Init() tea.Cmd {
	return func() tea.Msg {
		return reviewQuestionMsg{m.review}
	}
}

// UPDATE

type (
	reviewQuestionMsg struct {
		flashcard.Review
	}

	reviewFinishedMsg struct {
		flashcard.Review
	}
)

// nolint:cyclop
func (m reviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case viewportMsg:
		m.viewport = msg.viewport
		m.help.Width = m.viewport.Width
		return m, nil

	case reviewQuestionMsg:
		m.review = msg.Review
		m.status = reviewQuestion
		m.keys.skip.SetEnabled(m.review.Current() != m.review.Total())
		m.keys.answer.SetEnabled(true)
		m.keys.again.SetEnabled(false)
		m.keys.hard.SetEnabled(false)
		m.keys.normal.SetEnabled(false)
		m.keys.easy.SetEnabled(false)
		m.keys.veryEasy.SetEnabled(false)
		return m, nil

	case reviewFinishedMsg:
		m.review = msg.Review
		m.status = reviewFinished
		m.keys.answer.SetEnabled(false)
		m.keys.again.SetEnabled(false)
		m.keys.hard.SetEnabled(false)
		m.keys.normal.SetEnabled(false)
		m.keys.easy.SetEnabled(false)
		m.keys.veryEasy.SetEnabled(false)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.quit):
			if m.status == reviewFinished {
				return m, showDecks
			}
			return m, showCards(m.review.Deck(), firstCard)

		case key.Matches(msg, m.keys.answer):
			if m.status == reviewQuestion {
				m.status = reviewAnswer
				m.keys.skip.SetEnabled(false)
				m.keys.answer.SetEnabled(false)
				m.keys.again.SetEnabled(true)
				m.keys.hard.SetEnabled(true)
				m.keys.normal.SetEnabled(true)
				m.keys.easy.SetEnabled(true)
				m.keys.veryEasy.SetEnabled(true)
				return m, nil
			}

		case key.Matches(msg, m.keys.skip):
			if m.status == reviewQuestion && m.review.Total() > 1 {
				return m, skipCard(m.review)
			}

		case key.Matches(msg, m.keys.score):
			if m.status == reviewAnswer {
				return m, scoreCard(msg.String(), m.review, m.repository)
			}

		case key.Matches(msg, m.keys.showFullHelp):
			fallthrough
		case key.Matches(msg, m.keys.closeFullHelp):
			m.help.ShowAll = !m.help.ShowAll
		}
	}

	return m, nil
}

func skipCard(review flashcard.Review) tea.Cmd {
	return func() tea.Msg {
		review, err := review.Skip()
		if err != nil {
			return failed(err)
		}
		return reviewQuestionMsg{review}
	}
}

func scoreCard(input string, review flashcard.Review, repository *flashcard.Repository) tea.Cmd {
	return func() tea.Msg {
		score, err := flashcard.NewReviewScore(input)
		if err != nil {
			return failed(err)
		}

		stats, review, err := review.Rate(score)
		if err != nil {
			return failed(err)
		}

		if err = repository.Deck.Save(review.Deck()); err != nil {
			return failed(err)
		}

		if stats != nil {
			if err := repository.Stats.Save(review.Deck(), stats); err != nil {
				return failed(err)
			}
		}

		if review.Left() == 0 {
			return reviewFinishedMsg{review}
		}

		return reviewQuestionMsg{review}
	}
}

// VIEW

func (m reviewModel) View() string {
	switch m.status {
	case reviewQuestion:
		return reviewQuestionView(m)

	case reviewAnswer:
		return reviewAnswerView(m)

	case reviewFinished:
		return reviewFinishedView(m)

	default:
		return ""
	}
}

func reviewQuestionView(m reviewModel) string {
	content := titleReviewStyle.Render("Question")
	content += deckName.Render(m.review.Deck().Name)
	content += status.Render(fmt.Sprintf("%d of %d", m.review.Current(), m.review.Total()))

	card, err := m.review.CurrentCard()
	if err != nil {
		return errorView(err.Error())
	}

	markdown, err := RenderMarkdown(card.Question, m.viewport.Width)
	if err != nil {
		return errorView(err.Error())
	}

	content += markdownStyle.Render(markdown)
	content += helpReviewStyle.Render(m.help.View(m.keys))

	return reviewScreenStyle.Render(content)
}

func reviewAnswerView(m reviewModel) string {
	content := titleReviewStyle.Render("Answer")
	content += deckName.Render(m.review.Deck().Name)
	content += status.Render(fmt.Sprintf("%d of %d", m.review.Current(), m.review.Total()))

	card, err := m.review.CurrentCard()
	if err != nil {
		return errorView(err.Error())
	}

	markdown, err := RenderMarkdown(card.Answer, m.viewport.Width)
	if err != nil {
		return errorView(err.Error())
	}

	content += markdownStyle.Render(markdown)
	content += helpReviewStyle.Render(m.help.View(m.keys))

	return reviewScreenStyle.Render(content)

}

func reviewFinishedView(m reviewModel) string {
	total := m.review.Completed()
	content := titleStyle.Render("Congratulations!")
	content += normalTextStyle.Render(fmt.Sprintf("%d card%s reviewed.", total, pluralize(total, "s")))

	content += helpStyle.Render(m.help.View(m.keys))
	return largePaddingStyle.Render(content)
}
