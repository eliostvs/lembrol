package tui_test

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dustin/go-humanize"
	"github.com/stretchr/testify/assert"

	clock "github.com/eliostvs/lembrol/internal/clock/test"
	"github.com/eliostvs/lembrol/internal/tui"
)

func TestCardsList(t *testing.T) {
	t.Parallel()

	t.Run(
		"shows deck with no cards", func(t *testing.T) {
			view := newTestModel(t, emptyDeck).
				Init().
				SendKeyType(tea.KeyEnter).
				Get().
				View()

			assert.Contains(t, view, "Empty")
			assert.Contains(t, view, "No items.")
			assert.Contains(t, view, "a add • q quit • ? more")
		},
	)

	t.Run(
		"shows full help with card page without cards", func(t *testing.T) {
			view := newTestModel(t, emptyDeck).
				Init().
				SendKeyType(tea.KeyEnter).
				SendKeyRune(helpKey).
				Get().
				View()

			assert.Contains(t, view, "a add    q quit")
			assert.Contains(t, view, "? close help")
		},
	)

	t.Run(
		"shows deck with many cards", func(t *testing.T) {
			view := newTestModel(t, fewDecks).
				Init().
				SendKeyType(tea.KeyEnter).
				Get().
				View()

			assert.Contains(t, view, "Golang A")
			assert.Contains(t, view, "6 items")
			assert.Contains(t, view, activePrompt+latestCard.Question)
			assert.Contains(t, view, fmt.Sprintf("%sLast review %s", activePrompt, humanize.Time(latestCard.ReviewedAt)))
			assert.Contains(t, view, secondLatestCard.Question)
			assert.Contains(t, view, fmt.Sprintf("Last review %s", humanize.Time(secondLatestCard.ReviewedAt)))
			assert.Contains(t, view, secondLatestCard.Question)
			assert.Contains(t, view, fmt.Sprintf("Last review %s", humanize.Time(secondLatestCard.ReviewedAt)))
			assert.Contains(t, view, "••")
			assert.Contains(t, view, "↑/k up • ↓/j down • / filter • a add • s study • q quit • ? more")
		},
	)

	t.Run(
		"shows full help when there are many cards", func(t *testing.T) {
			view := newTestModel(t, manyDecks).
				Init().
				SendKeyType(tea.KeyEnter).
				SendKeyRune(helpKey).
				Get().
				View()

			assert.Contains(t, view, "↑/k      up             /     filter    q quit")
			assert.Contains(t, view, "↓/j      down           a     add       ? close help")
			assert.Contains(t, view, "→/l/pgdn next page      e     edit")
			assert.Contains(t, view, " ←/h/pgup prev page      x     delete")
			assert.Contains(t, view, "g/home   go to start    enter stats")
			assert.Contains(t, view, "G/end    go to end      s     study")
		},
	)

	t.Run(
		"truncates very long question text", func(t *testing.T) {
			view := newTestModel(t, longNamesDeck).
				Init().
				SendKeyType(tea.KeyEnter).
				Get().
				View()

			assert.Contains(t, view, "Very Long Question & Answer")
			assert.Contains(
				t,
				view,
				"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labo…",
			)
		},
	)

	t.Run(
		"shows homepage when deck is closed", func(t *testing.T) {
			view := newTestModel(t, singleCardDeck).
				Init().
				SendKeyType(tea.KeyEnter).
				SendKeyType(tea.KeyEsc).
				Get().
				View()

			assert.Contains(t, view, "Decks")
		},
	)

	t.Run(
		"navigates around", func(t *testing.T) {
			tests := []struct {
				name string
				keys []string
				want string
			}{
				{
					name: "moves up, down using arrow keyMap",
					keys: []string{tea.KeyDown.String(), tea.KeyUp.String()},
					want: activePrompt + latestCard.Question,
				},
				{
					name: "moves down, up, up",
					keys: []string{tea.KeyDown.String(), keyDown, keyDown},
					want: activePrompt + "Question D",
				},
				{
					name: "moves right, left using arrow keyMap",
					keys: []string{tea.KeyRight.String(), tea.KeyLeft.String()},
					want: activePrompt + latestCard.Question,
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name, func(t *testing.T) {
						var batch []tea.Msg
						for _, key := range tt.keys {
							batch = append(batch, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
						}

						view := newTestModel(t, manyDecks).
							Init().
							SendKeyType(tea.KeyEnter).
							SendBatch(batch).
							Get().
							View()

						assert.Contains(t, view, tt.want)
					},
				)
			}
		},
	)

	t.Run(
		"does not start review when it does not have due cards", func(t *testing.T) {
			view := newTestModel(t, singleCardDeck, tui.WithClock(clock.Clock{Time: latestCard.ReviewedAt})).
				Init().
				SendKeyType(tea.KeyEnter).
				SendKeyRune(studyKey).
				Get().
				View()

			assert.Contains(t, view, "Golang One")
		},
	)

	t.Run(
		"starts review", func(t *testing.T) {
			view := newTestModel(t, singleCardDeck).
				Init().
				SendKeyType(tea.KeyEnter).
				SendKeyRune(studyKey).
				Get().
				View()

			assert.Contains(t, view, latestCard.Question)
		},
	)

	t.Run(
		"filters cards", func(t *testing.T) {
			newTestModel(t, manyDecks).
				Init().
				SendKeyType(tea.KeyEnter).
				SendKeyRune(filterKey).
				SendKeyRune("C").
				Peek(
					func(m tea.Model) {
						view := m.View()
						assert.NotContains(t, view, "a add")
						assert.Contains(t, view, "Filter: C")
						assert.Contains(t, view, "1 item • 5 filtered")
					},
				).
				SendKeyType(tea.KeyEnter).
				Peek(
					func(m tea.Model) {
						view := m.View()
						assert.NotContains(t, view, "a add")
						assert.Contains(t, view, "“C” 1 item • 5 filtered")
					},
				)
		},
	)

	t.Run(
		"changes the layout when the window resize", func(t *testing.T) {
			view := newTestModel(t, manyDecks, tui.WithWindowSize(0, 0)).
				Init().
				SendKeyType(tea.KeyEnter).
				SendMsg(tea.WindowSizeMsg{Width: testWidth, Height: testHeight}).
				Get().
				View()

			assert.Contains(t, view, "Question C")
		},
	)
}

func TestCardAdd(t *testing.T) {
	t.Parallel()

	t.Run(
		"shows add card form", func(t *testing.T) {
			view := newTestModel(t, singleCardDeck).
				Init().
				SendKeyType(tea.KeyEnter).
				SendKeyRune(createKey).
				Get().
				View()

			assert.Contains(t, view, "Golang One")
			assert.Contains(t, view, "Add")
			assert.Contains(t, view, "nter a question")
			assert.Contains(t, view, "Enter an answer")
			assert.Contains(t, view, "↓ down • ↑ up • ctrl+s confirm • esc cancel")
		},
	)

	t.Run(
		"goes back to cards list when the creation is canceled", func(t *testing.T) {
			view := newTestModel(t, manyDecks).
				Init().
				SendKeyType(tea.KeyEnter).
				SendKeyType(tea.KeyDown).
				SendKeyRune(createKey).
				SendKeyType(tea.KeyEsc).
				Get().
				View()

			assert.Contains(t, view, activePrompt+"Question B")
			assert.NotContains(t, view, "enter confirm • esc cancel")
		},
	)

	t.Run(
		"shows deck when card is created", func(t *testing.T) {
			var showLoading bool

			assertCardCreated := func(view string) {
				assert.Contains(t, view, "Empty")
				assert.Contains(t, view, "1 item")
				assert.Contains(t, view, activePrompt+"New question")
				assert.Contains(t, view, activePrompt+"Last review just now")
				assert.Contains(t, view, "s study")
			}

			newTestModel(t, emptyDeck).
				WithObserver(
					func(m tea.Model) {
						if strings.Contains(m.View(), "Creating card...") {
							showLoading = true
						}
					},
				).
				Init().
				SendKeyType(tea.KeyEnter).
				SendKeyRune(createKey).
				SendKeyRune("New").
				SendKeyType(tea.KeyDown).
				SendKeyRune("New").
				SendKeyType(tea.KeyShiftTab).
				SendKeyRune(" question").
				SendKeyType(tea.KeyDown).
				SendKeyRune(" answer").
				SendKeyType(tea.KeyUp).
				SendKeyRune("?").
				SendKeyRune(saveKey).
				Peek(
					func(m tea.Model) {
						assertCardCreated(m.View())
						assert.True(t, showLoading)
					},
				).
				// confirm the state was correctly committed
				SendKeyType(tea.KeyEnter).
				SendKeyRune(quitKey).
				Peek(
					func(m tea.Model) {
						assertCardCreated(m.View())
					},
				)
		},
	)

	t.Run(
		"shows error when card the creation fail", func(t *testing.T) {
			view := newTestModel(t, errorDeck).
				Init().
				SendKeyType(tea.KeyEnter).
				SendKeyRune(createKey).
				SendKeyRune("Answer").
				SendKeyType(tea.KeyDown).
				SendKeyRune("Question").
				SendKeyType(tea.KeyEnter).
				Get().
				View()

			assert.Contains(t, view, "Error")
		},
	)

	t.Run(
		"doest not save card when", func(t *testing.T) {
			t.Run(
				"the question is empty", func(t *testing.T) {
					view := newTestModel(t, emptyDeck).
						Init().
						SendKeyType(tea.KeyEnter).
						SendKeyRune(createKey).
						SendKeyType(tea.KeyDown).
						SendKeyRune("Answer").
						SendKeyType(tea.KeyEnter).
						Get().
						View()

					assert.Contains(t, view, "Enter a question")
				},
			)

			t.Run(
				"the answer is empty", func(t *testing.T) {
					view := newTestModel(t, emptyDeck).
						Init().
						SendKeyType(tea.KeyEnter).
						SendKeyRune(createKey).
						SendKeyRune("Question").
						SendKeyType(tea.KeyEnter).
						Get().
						View()

					assert.Contains(t, view, "Enter an answer")
				},
			)
		},
	)
}

func TestCardEdit(t *testing.T) {
	t.Parallel()

	t.Run(
		"shows edit card form", func(t *testing.T) {
			view := newTestModel(t, singleCardDeck).
				Init().
				SendKeyType(tea.KeyEnter).
				SendKeyRune(editKey).
				Get().
				View()

			assert.Contains(t, view, "Golang One")
			assert.Contains(t, view, "Edit")
			assert.Contains(t, view, "┃ "+latestCard.Question)
			assert.Contains(t, view, "┃ "+latestCard.Answer)
			assert.Contains(t, view, "↓ down • ↑ up • ctrl+s confirm • esc cancel")
		},
	)

	t.Run(
		"goes back to card list when edition is canceled", func(t *testing.T) {
			view := newTestModel(t, manyDecks).
				Init().
				SendKeyType(tea.KeyEnter).
				SendKeyType(tea.KeyDown).
				SendKeyRune(editKey).
				SendKeyType(tea.KeyEsc).
				Get().
				View()

			assert.Contains(t, view, activePrompt+"Question B")
			assert.NotContains(t, view, "enter confirm • esc cancel")
		},
	)

	t.Run(
		"shows error when the edition fail", func(t *testing.T) {
			view := newTestModel(t, errorDeck).
				Init().
				SendKeyType(tea.KeyEnter).
				SendKeyRune(editKey).
				SendKeyRune(saveKey).
				Get().
				View()

			assert.Contains(t, view, "Error")
		},
	)

	t.Run(
		"shows deck when card is changed", func(t *testing.T) {
			var showLoading bool

			assertCardChanged := func(view string) {
				assert.Contains(t, view, activePrompt+latestCard.Question+"-q")
			}

			newTestModel(t, manyDecks).
				WithObserver(
					func(m tea.Model) {
						if strings.Contains(m.View(), "Updating card...") {
							showLoading = true
						}
					},
				).
				Init().
				SendKeyType(tea.KeyEnter).
				SendKeyRune(editKey).
				SendKeyRune("--").
				SendKeyType(tea.KeyDown).
				SendKeyRune("--").
				SendKeyType(tea.KeyUp).
				SendKeyType(tea.KeyBackspace).
				SendKeyType(tea.KeyTab).
				SendKeyType(tea.KeyBackspace).
				SendKeyType(tea.KeyShiftTab).
				SendKeyRune("q").
				SendKeyType(tea.KeyTab).
				SendKeyRune("q").
				SendKeyRune(saveKey).
				Peek(
					func(m tea.Model) {
						assertCardChanged(m.View())
						assert.True(t, showLoading)
					},
				).
				// confirm the state was correctly committed
				SendKeyType(tea.KeyEnter).
				SendKeyRune(quitKey).
				Peek(
					func(m tea.Model) {
						assertCardChanged(m.View())
					},
				)
		},
	)
}

func TestCardDelete(t *testing.T) {
	t.Parallel()

	t.Run(
		"confirms card deletion", func(t *testing.T) {
			view := newTestModel(t, fewDecks).
				Init().
				SendKeyType(tea.KeyEnter).
				SendKeyRune(deleteKey).
				Get().
				View()

			assert.Contains(t, view, "Delete this card?")
			assert.Contains(t, view, activePrompt+latestCard.Question)
			assert.Contains(t, view, fmt.Sprintf("%sLast review %s", activePrompt, humanize.Time(latestCard.ReviewedAt)))
		},
	)

	t.Run(
		"shows cards when the card deletion is canceled", func(t *testing.T) {
			view := newTestModel(t, manyDecks).
				Init().
				SendKeyType(tea.KeyEnter).
				SendKeyRune(deleteKey).
				SendKeyType(tea.KeyEsc).
				Get().
				View()

			assert.NotContains(t, view, "Delete this card?")
			assert.Contains(t, view, activePrompt+latestCard.Question)
		},
	)

	t.Run(
		"ignores inputs besides cancel or enter", func(t *testing.T) {
			tests := []struct {
				name string
				key  string
			}{
				{
					name: "delete",
					key:  deleteKey,
				},
				{
					name: "create",
					key:  createKey,
				},
				{
					name: "study",
					key:  studyKey,
				},
				{
					name: "edit",
					key:  editKey,
				},
				{
					name: "help",
					key:  helpKey,
				},
				{
					name: "down",
					key:  "down",
				},
				{
					name: "up",
					key:  "up",
				},
				{
					name: "left",
					key:  "left",
				},
				{
					name: "right",
					key:  "right",
				},
				{
					name: "home",
					key:  "home",
				},
				{
					name: "down",
					key:  "down",
				},
				{
					name: "filter",
					key:  filterKey,
				},
			}
			for _, tt := range tests {
				t.Run(
					tt.name, func(t *testing.T) {
						m := newTestModel(t, manyDecks).
							Init().
							SendKeyType(tea.KeyEnter).
							SendKeyRune(deleteKey).
							SendKeyRune(tt.key).
							Get()

						view := m.View()

						assert.Contains(t, view, "Delete this card?")
					},
				)
			}
		},
	)

	t.Run(
		"ignores delete action when deck has no cards", func(t *testing.T) {
			m := newTestModel(t, noneDeck).
				Init().
				SendKeyType(tea.KeyEnter).
				SendKeyRune(deleteKey).
				Get()

			assert.NotContains(t, m.View(), "Delete this card?")
		},
	)

	t.Run(
		"shows error when card deletion fails", func(t *testing.T) {
			var showLoading bool

			m := newTestModel(t, errorDeck).
				WithObserver(
					func(m tea.Model) {
						if strings.Contains(m.View(), "Deleting card...") {
							showLoading = true
						}
					},
				).
				Init().
				SendKeyType(tea.KeyEnter).
				SendKeyRune(deleteKey).
				SendKeyType(tea.KeyEnter).
				Get()

			assert.Contains(t, m.View(), "Error")
			assert.True(t, showLoading)
		},
	)

	t.Run(
		"shows deck when card is deleted", func(t *testing.T) {
			assertCardDeleted := func(view string) {
				assert.NotContains(t, view, "Delete this card?")
				assert.NotContains(t, view, latestCard.Question)
			}

			newTestModel(t, fewDecks).
				Init().
				SendKeyType(tea.KeyEnter).
				SendKeyRune(deleteKey).
				SendKeyType(tea.KeyEnter).
				Peek(
					func(m tea.Model) {
						assertCardDeleted(m.View())
					},
				).
				SendKeyType(tea.KeyEnter).
				SendKeyType(tea.KeyEsc).
				Peek(
					func(m tea.Model) {
						assertCardDeleted(m.View())
					},
				)
		},
	)
}
