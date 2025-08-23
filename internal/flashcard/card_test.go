package flashcard_test

import (
	"testing"
	"time"

	fsrs "github.com/open-spaced-repetition/go-fsrs/v3"
	"github.com/stretchr/testify/assert"

	"github.com/eliostvs/lembrol/internal/flashcard"
)

func TestCard_IsDue(t *testing.T) {
	t.Parallel()

	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)
	yesterday := now.Add(-24 * time.Hour)

	tests := []struct {
		name string
		card flashcard.Card
		want bool
	}{
		{
			name: "returns true when card is due now",
			card: flashcard.Card{Due: now},
			want: true,
		},
		{
			name: "returns true when card was due yesterday",
			card: flashcard.Card{Due: yesterday},
			want: true,
		},
		{
			name: "returns false when card is due tomorrow",
			card: flashcard.Card{Due: tomorrow},
			want: false,
		},
		{
			name: "returns true for new cards with zero due date",
			card: flashcard.Card{LastReview: now},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.card.IsDue(now))
		})
	}
}

func TestNewCard(t *testing.T) {
	t.Parallel()

	now := time.Now()
	question := "What is FSRS?"
	answer := "Free Spaced Repetition Scheduler"

	card := flashcard.NewCard(question, answer, now)

	assert.NotEmpty(t, card.ID)
	assert.Equal(t, question, card.Question)
	assert.Equal(t, answer, card.Answer)
	assert.Equal(t, now, card.LastReview)
	assert.Equal(t, now, card.Due)
	assert.Equal(t, now, card.LastReview)
	assert.Zero(t, card.Stability)
	assert.Zero(t, card.Difficulty)
	assert.Equal(t, fsrs.New, card.State)
	assert.Zero(t, card.Reps)
	assert.Zero(t, card.Lapses)
}

func TestReviewScore_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		score flashcard.ReviewScore
		want  string
	}{
		{
			name:  "score again",
			score: flashcard.ReviewScoreAgain,
			want:  "0",
		},
		{
			name:  "score hard",
			score: flashcard.ReviewScoreHard,
			want:  "1",
		},
		{
			name:  "score normal",
			score: flashcard.ReviewScoreNormal,
			want:  "2",
		},
		{
			name:  "score easy",
			score: flashcard.ReviewScoreEasy,
			want:  "3",
		},
		{
			name:  "score super easy",
			score: flashcard.ReviewScoreSuperEasy,
			want:  "4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.score.String())
		})
	}
}

func TestNewReviewScore(t *testing.T) {
	t.Parallel()

	type want struct {
		err   error
		score flashcard.ReviewScore
	}
	tests := []struct {
		name string
		args string
		want
	}{
		{
			name: "score again",
			args: "0",
			want: want{score: flashcard.ReviewScoreAgain},
		},
		{
			name: "score hard",
			args: "1",
			want: want{score: flashcard.ReviewScoreHard},
		},
		{
			name: "score normal",
			args: "2",
			want: want{score: flashcard.ReviewScoreNormal},
		},
		{
			name: "score easy",
			args: "3",
			want: want{score: flashcard.ReviewScoreEasy},
		},
		{
			name: "score super easy",
			args: "4",
			want: want{score: flashcard.ReviewScoreSuperEasy},
		},
		{
			name: "fails number bigger than score super easy",
			args: "5",
			want: want{score: flashcard.ReviewScore(-1), err: flashcard.ErrInvalidScore},
		},
		{
			name: "fails when number smaller than score again",
			args: "-1",
			want: want{score: flashcard.ReviewScore(-1), err: flashcard.ErrInvalidScore},
		},
		{
			name: "fails when is not a number",
			args: "a",
			want: want{score: flashcard.ReviewScore(-1), err: flashcard.ErrInvalidScore},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := flashcard.NewReviewScore(tt.args)

			assert.Equal(t, tt.want.err, err)
			assert.Equal(t, tt.want.score, got)
		})
	}
}

/*
 Test Utilities
*/

var (
	oldestCard = flashcard.Card{
		Question:   "How do you delete a file?",
		Answer:     "import \"os\"\n\nos.Remove(path) error",
		LastReview: time.Date(2021, 1, 2, 15, 0, 0, 0, time.UTC),
		Due:        time.Date(2021, 1, 2, 15, 0, 0, 0, time.UTC),
		State:      fsrs.New,
		Stability:  1.0,
		Difficulty: 5.0,
	}
	afterOldestCard  = oldestCard.LastReview.Add(24 * time.Hour)
	beforeOldestCard = oldestCard.LastReview.Add(-24 * time.Hour)
)
