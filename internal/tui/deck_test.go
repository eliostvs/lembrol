package tui_test

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	clock "github.com/eliostvs/lembrol/internal/clock/test"
	"github.com/eliostvs/lembrol/internal/tui"
)

func TestDecksList(t *testing.T) {
	t.Parallel()

	t.Run(
		"shows homepage without decks", func(t *testing.T) {
			view := newTestModel(t, noneDeck).
				Init().
				Get().
				View()

			assert.NotContains(t, view, "↑/k up")
			assert.NotContains(t, view, "↓/j down")
			assert.NotContains(t, view, "/ filter")
			assert.NotContains(t, view, "enter open")
			assert.Contains(t, view, "No items.")
			assert.Contains(t, view, "a add • q quit • ? more")
		},
	)

	t.Run(
		"shows fulls help when there are no decks", func(t *testing.T) {
			view := newTestModel(t, noneDeck).
				Init().
				SendKeyRune(helpKey).
				Get().
				View()

			assert.Contains(t, view, "a add    q quit")
			assert.Contains(t, view, "? close help")
		},
	)

	t.Run(
		"shows homepage with many decks", func(t *testing.T) {
			m := newTestModel(t, manyDecks, tui.WithClock(clock.New(oldestCard.LastReview.Add(48*time.Hour)))).
				Init().
				Get()

			view := m.View()

			assert.Contains(t, view, "Decks")
			assert.Contains(t, view, "6 items")
			assert.Contains(t, view, activePrompt+"Golang A")
			assert.Contains(t, view, activePrompt+"6 cards | 3 due")
			assert.Contains(t, view, "Golang B")
			assert.Contains(t, view, "3 cards | 0 due")
			assert.NotContains(t, view, activePrompt+"Golang B")
			assert.Contains(t, view, "••")
			assert.Contains(t, view, "↑/k up • ↓/j down • / filter • a add • enter open • q quit • ? more")
		},
	)

	t.Run(
		"shows full help when there are many decks", func(t *testing.T) {
			view := newTestModel(t, manyDecks).
				Init().
				SendKeyRune(helpKey).
				Get().
				View()

			assert.Contains(t, view, "↑/k      up             /     filter    q quit")
			assert.Contains(t, view, "↓/j      down           a     add       ? close help")
			assert.Contains(t, view, "→/l/pgdn next page      enter open")
			assert.Contains(t, view, "g/home   go to start    x     delete")
			assert.Contains(t, view, "G/end    go to end      s     study")
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
					name: "using arrow keys 1",
					keys: []string{tea.KeyDown.String()},
					want: activePrompt + "Golang B",
				},
				{
					name: "using vim keys 1",
					keys: []string{keyDown, keyDown, keyDown},
					want: activePrompt + "Golang D",
				},
				{
					name: "using arrow keys 2",
					keys: []string{tea.KeyDown.String(), tea.KeyUp.String(), tea.KeyUp.String()},
					want: activePrompt + "Golang A",
				},
				{
					name: "using vim keys 2",
					keys: []string{keyDown, keyUp, keyUp},
					want: activePrompt + "Golang A",
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
		"quits the app", func(t *testing.T) {
			view := newTestModel(t, noneDeck).
				Init().
				SendKeyRune(quitKey).
				Get().
				View()

			assert.Contains(t, view, "Thanks for using Lembrol!")
		},
	)

	t.Run(
		"ignores actions when there are no decks", func(t *testing.T) {
			tests := []struct {
				name string
				key  string
			}{
				{
					name: "open deck",
					key:  tea.KeyEnter.String(),
				},
				{
					name: "starts review",
					key:  studyKey,
				},
				{
					name: "deletes deck",
					key:  studyKey,
				},
				{
					name: "chooses deck above",
					key:  keyUp,
				},
				{
					name: "chooses deck bellow",
					key:  keyDown,
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name, func(t *testing.T) {
						view := newTestModel(t, noneDeck).
							Init().
							SendKeyRune(tt.key).
							Get().
							View()

						assert.Contains(t, view, "No items")
					},
				)
			}
		},
	)

	t.Run(
		"does not start review when it does not have due cards", func(t *testing.T) {
			view := newTestModel(t, manyDecks, tui.WithClock(clock.Clock{Time: oldestCard.LastReview.Add(-time.Hour)})).
				Init().
				SendKeyRune(studyKey).
				Get().
				View()

			assert.Contains(t, view, "Decks")
			assert.NotContains(t, view, "Question")
		},
	)

	t.Run(
		"shows questions when review starts", func(t *testing.T) {
			view := newTestModel(t, singleCardDeck).
				Init().
				SendKeyRune(studyKey).
				Get().
				View()

			assert.Contains(t, view, latestCard.Question)
		},
	)

	t.Run(
		"shows deck when it is open", func(t *testing.T) {
			view := newTestModel(t, singleCardDeck).
				Init().
				SendKeyType(tea.KeyEnter).
				Get().
				View()

			assert.Contains(t, view, "Golang One")
			assert.NotContains(t, view, "Decks")
		},
	)

	t.Run(
		"filters decks", func(t *testing.T) {
			newTestModel(t, manyDecks).
				Init().
				SendKeyRune(filterKey).
				SendKeyRune("B").
				Peek(
					func(m tea.Model) {
						view := m.View()
						assert.NotContains(t, view, "a add")
						assert.Contains(t, view, "Filter: B")
						assert.Contains(t, view, "1 item • 5 filtered")
					},
				).
				SendKeyType(tea.KeyEnter).
				Peek(
					func(m tea.Model) {
						view := m.View()
						assert.NotContains(t, view, "a add")
						assert.Contains(t, view, "“B” 1 item • 5 filtered")
					},
				)
		},
	)

	t.Run(
		"changes the layout when the window resize", func(t *testing.T) {
			view := newTestModel(t, manyDecks, tui.WithWindowSize(0, 0)).
				Init().
				SendMsg(tea.WindowSizeMsg{Width: testWidth, Height: testHeight}).
				Get().
				View()

			assert.Contains(t, view, "Golang C")
		},
	)
}

func TestDeckCreate(t *testing.T) {
	t.Parallel()

	t.Run(
		"shows crate form", func(t *testing.T) {
			view := newTestModel(t, noneDeck).
				Init().
				SendKeyRune(createKey).
				Get().
				View()

			assert.Contains(t, view, "Add")
			assert.Contains(t, view, "ctrl+s confirm • esc cancel")
		},
	)

	t.Run(
		"shows homepage when the creation is canceled", func(t *testing.T) {
			view := newTestModel(t, manyDecks).
				Init().
				SendKeyType(tea.KeyDown).
				SendKeyRune(createKey).
				SendKeyType(tea.KeyEsc).
				Get().
				View()

			assert.Contains(t, view, "Decks")
			assert.Contains(t, view, activePrompt+"Golang B")
			assert.NotContains(t, view, "Add")
		},
	)

	t.Run(
		"validates deck name can't be empty", func(t *testing.T) {
			view := newTestModel(t, noneDeck).
				Init().
				SendKeyRune(createKey).
				SendKeyType(tea.KeyEnter).
				Get().
				View()

			assert.Contains(t, view, "Add")
		},
	)

	t.Run(
		"shows homepage when deck is created", func(t *testing.T) {
			var showLoading bool

			newTestModel(t, noneDeck).
				WithObserver(
					func(m tea.Model) {
						if strings.Contains(m.View(), "Creating deck...") {
							showLoading = true
						}
					},
				).
				Init().
				SendKeyRune(createKey).
				SendKeyRune("Golang q").
				SendKeyRune(saveKey).
				Peek(
					func(m tea.Model) {
						view := m.View()
						assert.Contains(t, view, "Decks")
						assert.Contains(t, view, "1 item")
						assert.True(t, showLoading)
					},
				).
				SendKeyType(tea.KeyEnter).
				SendKeyType(tea.KeyEsc).
				Peek(
					func(m tea.Model) {
						view := m.View()
						assert.Contains(t, view, "Decks")
						assert.Contains(t, view, "1 item")
					},
				)
		},
	)

	t.Run(
		"shows error when create deck fail", func(t *testing.T) {
			view := newTestModel(t, t.TempDir()).
				Init().
				SendKeyRune(createKey).
				SendKeyRune("Error").
				SendKeyRune(saveKey).
				Get().
				View()

			assert.Contains(t, view, "Error")
		},
	)
}

func TestDeckEdit(t *testing.T) {
	t.Parallel()

	t.Run(
		"shows edit form", func(t *testing.T) {
			view := newTestModel(t, fewDecks).
				Init().
				SendKeyRune(editKey).
				Get().
				View()

			assert.Contains(t, view, "Edit")
			assert.Contains(t, view, "ctrl+s confirm • esc cancel")
		},
	)

	t.Run(
		"shows homepage when the edit is canceled", func(t *testing.T) {
			view := newTestModel(t, manyDecks).
				Init().
				SendKeyType(tea.KeyDown).
				SendKeyRune(editKey).
				SendKeyType(tea.KeyEsc).
				Get().
				View()

			assert.Contains(t, view, "Decks")
			assert.Contains(t, view, activePrompt+"Golang B")
			assert.NotContains(t, view, "Edit")
		},
	)

	t.Run(
		"shows homepage when deck is edited", func(t *testing.T) {
			var showLoading bool

			assertDeckChanged := func(view string) {
				assert.Contains(t, view, "Decks")
				assert.NotContains(t, view, "Rename this deck?")
				assert.Contains(t, view, activePrompt+"Golang AA!")
			}

			newTestModel(t, fewDecks).
				WithObserver(
					func(m tea.Model) {
						if strings.Contains(m.View(), "Saving deck...") {
							showLoading = true
						}
					},
				).
				Init().
				SendKeyRune(editKey).
				SendKeyType(tea.KeyBackspace).
				SendKeyRune("AA!").
				SendKeyRune(saveKey).
				Peek(
					func(m tea.Model) {
						assertDeckChanged(m.View())
						assert.True(t, showLoading)
					},
				).
				SendKeyType(tea.KeyEnter).
				SendKeyType(tea.KeyEsc).
				Peek(
					func(m tea.Model) {
						assertDeckChanged(m.View())
						assert.True(t, showLoading)
					},
				)
		},
	)

	t.Run(
		"shows error when edition fail", func(t *testing.T) {
			view := newTestModel(t, errorDeck).
				Init().
				SendKeyRune(editKey).
				SendKeyRune(" Change").
				SendKeyRune(saveKey).
				Get().
				View()

			assert.Contains(t, view, "Error")
		},
	)
}

func TestDeckDelete(t *testing.T) {
	t.Parallel()

	t.Run(
		"confirms the deletion", func(t *testing.T) {
			view := newTestModel(t, fewDecks).
				Init().
				SendKeyRune(tea.KeyDown.String()).
				SendKeyRune(deleteKey).
				Get().
				View()

			assert.Contains(t, view, "Delete this deck?")
			assert.Contains(t, view, activePrompt+"Golang B")
			assert.Contains(t, view, activePrompt+"2 cards | 2 due")
			assert.Contains(t, view, "enter confirm • esc cancel")
		},
	)

	t.Run(
		"shows homepage when the deletion is canceled", func(t *testing.T) {
			view := newTestModel(t, manyDecks).
				Init().
				SendKeyRune(deleteKey).
				SendKeyType(tea.KeyEsc).
				Get().
				View()

			assert.NotContains(t, view, "Delete this deck?")
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
						view := newTestModel(t, manyDecks).
							Init().
							SendKeyRune(deleteKey).
							SendKeyRune(tt.key).
							Get().
							View()

						assert.Contains(t, view, "Delete this deck?")
					},
				)
			}
		},
	)

	t.Run(
		"shows homepage when the deck is deleted", func(t *testing.T) {
			var showLoading bool

			view := newTestModel(t, manyDecks).
				WithObserver(
					func(m tea.Model) {
						if strings.Contains(m.View(), "Deleting deck...") {
							showLoading = true
						}
					},
				).
				Init().
				SendKeyRune(deleteKey).
				SendKeyType(tea.KeyEnter).
				Get().
				View()

			assert.Contains(t, view, "5 items")
			assert.Contains(t, view, "↑/k up • ↓/j down • / filter • a add • enter open • q quit • ? more")
			assert.True(t, showLoading)
		},
	)

	t.Run(
		"shows error when the deletion fail", func(t *testing.T) {
			view := newTestModel(t, errorDeck).
				Init().
				SendKeyRune(deleteKey).
				SendKeyType(tea.KeyEnter).
				Get().
				View()

			assert.Contains(t, view, "Error")
		},
	)
}
