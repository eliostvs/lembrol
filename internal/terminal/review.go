package terminal

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/eliostvs/lembrol/internal/flashcard"
)

// Messages

func showAnswer(review flashcard.Review) tea.Cmd {
	return func() tea.Msg {
		return showAnswerMsg{
			review,
		}
	}
}

func skipCard(review flashcard.Review) tea.Cmd {
	return func() tea.Msg {
		review, err := review.Skip()
		if err != nil {
			return fail(err)
		}
		return showQuestionMsg{review}
	}
}

func scoreCard(rawScore string, review flashcard.Review, repository Repository) tea.Cmd {
	return func() tea.Msg {
		score, err := flashcard.NewReviewScore(rawScore)
		if err != nil {
			return fail(err)
		}

		review, err := review.Rate(score)
		if err != nil {
			return fail(err)
		}

		if err = repository.Save(review.Deck); err != nil {
			return fail(err)
		}

		if review.Left() == 0 {
			return showReviewSummaryMsg{review}
		}

		return showQuestionMsg{review}
	}
}

type (
	showQuestionMsg struct {
		flashcard.Review
	}

	showAnswerMsg struct {
		flashcard.Review
	}

	showReviewSummaryMsg struct {
		flashcard.Review
	}

	setupQuestionMsg struct{}

	setupAnswerPageMsg struct{}
)

// Question Page

type questionKeyMap struct {
	skip, answer, quit key.Binding
}

func (k questionKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.skip, k.answer, k.quit}
}

func (k questionKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.skip,
		},
		{
			k.answer,
		},
		{
			k.quit,
		},
	}
}

func newQuestionPage(common reviewCommon) questionPage {
	return questionPage{
		reviewCommon: common,
		keyMap: questionKeyMap{
			answer: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "answer"),
			),
			skip: key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "skip"),
			),
			quit: key.NewBinding(
				key.WithKeys("q", "esc"),
				key.WithHelp("q", "quit"),
			),
		},
	}
}

type questionPage struct {
	reviewCommon
	keyMap questionKeyMap
}

func (m questionPage) Init() tea.Cmd {
	return func() tea.Msg {
		return setupQuestionMsg{}
	}
}

func (m questionPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case setupQuestionMsg:
		m.keyMap.skip.SetEnabled(m.review.Current() != m.review.Total())
		return m, nil

	case innerWindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.skip) && m.review.Total() > 1:
			return m, skipCard(m.review)

		case key.Matches(msg, m.keyMap.answer):
			return m, showAnswer(m.review)

		case key.Matches(msg, m.keyMap.quit):
			return m, showCards(0, m.review.Deck)
		}
	}

	return m, nil
}

func (m questionPage) View() string {
	var content strings.Builder

	content.WriteString(titleStyle.Render("Question"))
	content.WriteString(subTitleStyle.Render(m.review.Deck.Name))
	content.WriteString(paginationStyle.Render(fmt.Sprintf("%d of %d", m.review.Current(), m.review.Total())))

	card, err := m.review.Card()
	if err != nil {
		return errorView(err.Error())
	}

	markdown, err := RenderMarkdown(card.Question, m.width)
	if err != nil {
		return errorView(err.Error())
	}

	content.WriteString(markdown)
	content.WriteString(renderHelp(m.keyMap, m.width, m.height-lipgloss.Height(content.String()), false))

	return largePaddingStyle.Render(content.String())
}

// Answer Page

type answerKeyMap struct {
	quit, score, again, hard, normal, easy, veryEasy, showFullHelp, closeFullHelp key.Binding
}

func (k answerKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.hard, k.normal, k.easy, k.veryEasy, k.quit, k.showFullHelp}
}

func (k answerKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.again,
		},
		{
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

func newAnswerPage(common reviewCommon) answerPage {
	return answerPage{
		reviewCommon: common,
		keyMap: answerKeyMap{
			score: key.NewBinding(
				key.WithKeys("0", "1", "2", "3", "4"),
				key.WithHelp("1", "score"),
			),
			again: key.NewBinding(
				key.WithKeys("0"),
				key.WithHelp("0", "again"),
			),
			hard: key.NewBinding(
				key.WithKeys("1"),
				key.WithHelp("1", "hard"),
			),
			normal: key.NewBinding(
				key.WithKeys("2"),
				key.WithHelp("2", "normal"),
			),
			easy: key.NewBinding(
				key.WithKeys("3"),
				key.WithHelp("3", "easy"),
			),
			veryEasy: key.NewBinding(
				key.WithKeys("4"),
				key.WithHelp("4", "very easy"),
			),
			showFullHelp: key.NewBinding(
				key.WithKeys("?"),
				key.WithHelp("?", "more"),
			),
			closeFullHelp: key.NewBinding(
				key.WithKeys("?"),
				key.WithHelp("?", "close help"),
			),
			quit: key.NewBinding(
				key.WithKeys("q", "esc"),
				key.WithHelp("q", "quit"),
			),
		},
	}
}

type answerPage struct {
	reviewCommon
	keyMap   answerKeyMap
	fullHelp bool
}

func (m answerPage) Init() tea.Cmd {
	return func() tea.Msg {
		return setupAnswerPageMsg{}
	}
}

func (m answerPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case setupAnswerPageMsg:
		m.keyMap.again.SetEnabled(m.review.Total() > 1)
		return m, nil

	case innerWindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.score):
			return m, tea.Batch(showLoading("Review", "Scoring card..."), scoreCard(msg.String(), m.review, m.repository))

		case key.Matches(msg, m.keyMap.showFullHelp):
			fallthrough

		case key.Matches(msg, m.keyMap.closeFullHelp):
			m.fullHelp = !m.fullHelp
			return m, nil

		case key.Matches(msg, m.keyMap.quit):
			return m, showCards(0, m.review.Deck)
		}
	}

	return m, nil
}

func (m answerPage) View() string {
	var content strings.Builder

	content.WriteString(titleStyle.Render("Answer"))
	content.WriteString(subTitleStyle.Render(m.review.Deck.Name))
	content.WriteString(paginationStyle.Render(fmt.Sprintf("%d of %d", m.review.Current(), m.review.Total())))

	card, err := m.review.Card()
	if err != nil {
		return errorView(err.Error())
	}

	markdown, err := RenderMarkdown(card.Answer, m.width)
	if err != nil {
		return errorView(err.Error())
	}

	content.WriteString(markdown)
	content.WriteString(renderHelp(m.keyMap, m.width, m.height-lipgloss.Height(content.String()), m.fullHelp))
	return largePaddingStyle.Render(content.String())
}

// Review Summary Page

type reviewSummaryKeyMap struct {
	quit key.Binding
}

func (k reviewSummaryKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.quit}
}

func (k reviewSummaryKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.quit}}
}

func newReviewSummaryPage(common reviewCommon) reviewSummaryPage {
	return reviewSummaryPage{
		reviewCommon: common,
		keyMap: reviewSummaryKeyMap{
			quit: key.NewBinding(
				key.WithKeys("q", "esc"),
				key.WithHelp("q", "quit"),
			),
		},
	}
}

type reviewSummaryPage struct {
	reviewCommon
	keyMap reviewSummaryKeyMap
}

func (m reviewSummaryPage) Init() tea.Cmd {
	return nil
}

func (m reviewSummaryPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case innerWindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.quit):
			return m, showDecks(0)
		}
	}

	return m, nil
}

func (m reviewSummaryPage) View() string {
	completed := m.review.Completed

	var content strings.Builder
	content.WriteString(titleStyle.Render("Congratulations!"))
	content.WriteString(normalTextStyle.Render(fmt.Sprintf("%d card%s reviewed.", completed, pluralize(completed, "s"))))

	content.WriteString(renderHelp(m.keyMap, m.width, m.height-lipgloss.Height(content.String()), false))
	return largePaddingStyle.Render(content.String())
}

// Review SubPage

func newReviewPage(review flashcard.Review, parent appCommon) reviewPage {
	common := reviewCommon{
		appCommon: parent,
		review:    review,
	}

	return reviewPage{
		reviewCommon: common,
		page:         newQuestionPage(common),
	}
}

type reviewCommon struct {
	appCommon
	review flashcard.Review
}

type reviewPage struct {
	reviewCommon
	page tea.Model
}

func (m reviewPage) Init() tea.Cmd {
	return func() tea.Msg {
		return showQuestionMsg{m.review}
	}
}

func (m reviewPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case showLoadingMsg:
		m.page = newLoadingPage(msg.title, msg.description, m.appCommon)
		return m, m.page.Init()

	case showAnswerMsg:
		m.review = msg.Review
		m.page = newAnswerPage(m.reviewCommon)
		return m, m.page.Init()

	case showQuestionMsg:
		m.review = msg.Review
		m.page = newQuestionPage(m.reviewCommon)
		return m, m.page.Init()

	case showReviewSummaryMsg:
		m.review = msg.Review
		m.page = newReviewSummaryPage(m.reviewCommon)
		return m, m.page.Init()
	}

	m.page, cmd = m.page.Update(msg)
	return m, cmd
}

func (m reviewPage) View() string {
	return m.page.View()
}
