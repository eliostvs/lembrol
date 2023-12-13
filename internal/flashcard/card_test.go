package flashcard_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/eliostvs/lembrol/internal/flashcard"
)

func TestCard_Advance(t *testing.T) {
	t.Parallel()

	now := time.Now()
	tomorrow := now.Add(time.Hour * 24)

	type args struct {
		card  flashcard.Card
		score flashcard.ReviewScore
	}
	type want struct {
		repetitions    int
		nextReview     time.Time
		easinessFactor float64
	}
	tests := []struct {
		name string
		args
		want
	}{
		{
			name: "review score is again",
			args: args{
				card: flashcard.Card{
					Repetitions: 10,
				},
				score: flashcard.ReviewScoreAgain,
			},
			want: want{
				repetitions: 0,
				nextReview:  tomorrow,
			},
		},
		{
			name: "review score is hard",
			args: args{
				card: flashcard.Card{
					Repetitions: 10,
				},
				score: flashcard.ReviewScoreHard,
			},
			want: want{
				repetitions: 0,
				nextReview:  tomorrow,
			},
		},
		{
			name: "score is normal in the first repetition",
			args: args{
				card: flashcard.Card{
					Repetitions: 0,
				},
				score: flashcard.ReviewScoreNormal,
			},
			want: want{
				repetitions:    1,
				nextReview:     tomorrow,
				easinessFactor: 1.3,
			},
		},
		{
			name: "score is easy in the second repetition",
			args: args{
				card: flashcard.Card{
					Repetitions: 1,
				},
				score: flashcard.ReviewScoreEasy,
			},
			want: want{
				repetitions:    2,
				nextReview:     now.Add(time.Hour * 24 * 6),
				easinessFactor: 1.3,
			},
		},
		{
			name: "score is very easy in the third repetition",
			args: args{
				card: flashcard.Card{
					Repetitions:    2,
					Interval:       2,
					EasinessFactor: 2,
				},
				score: flashcard.ReviewScoreEasy,
			},
			want: want{
				repetitions:    3,
				nextReview:     now.Add(time.Hour * 24 * 4),
				easinessFactor: 2,
			},
		},
		{
			name: "score is normal lower the easiness factor",
			args: args{
				card: flashcard.Card{
					Repetitions:    2,
					Interval:       1,
					EasinessFactor: 2,
				},
				score: flashcard.ReviewScoreNormal,
			},
			want: want{
				repetitions:    3,
				nextReview:     now.Add(time.Hour * 24 * 2),
				easinessFactor: 1.9,
			},
		},
		{
			name: "score is easy keep the same easiness factor",
			args: args{
				card: flashcard.Card{
					Repetitions:    2,
					Interval:       1,
					EasinessFactor: 2,
				},
				score: flashcard.ReviewScoreEasy,
			},
			want: want{
				repetitions:    3,
				nextReview:     now.Add(time.Hour * 24 * 2),
				easinessFactor: 2,
			},
		},
		{
			name: "score is very easy rise the easiness factor",
			args: args{
				card: flashcard.Card{
					Repetitions:    2,
					Interval:       1,
					EasinessFactor: 2,
				},
				score: flashcard.ReviewScoreSuperEasy,
			},
			want: want{
				repetitions:    3,
				nextReview:     now.Add(time.Hour * 24 * 2),
				easinessFactor: 2.1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				card := tt.card.Advance(now, tt.score)

				assert.Equal(t, tt.want.repetitions, card.Repetitions)
				assert.Equal(t, tt.want.nextReview, card.NextReviewAt())
				assert.GreaterOrEqual(t, tt.want.easinessFactor, card.EasinessFactor)
				assert.Equal(t, []flashcard.Stats{flashcard.NewStats(now, tt.score, tt.card)}, card.Stats)
			},
		)
	}
}

func TestCard_Due(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name string
		card flashcard.Card
		want bool
	}{
		{
			name: "returns true when card is due",
			card: flashcard.Card{ReviewedAt: now.Add(-time.Hour)},
			want: true,
		},
		{
			name: "returns true when card is due",
			card: flashcard.Card{ReviewedAt: now.Add(time.Hour)},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				assert.Equal(t, tt.card.IsDue(now), tt.want)
			},
		)
	}
}

func TestReviewScore_String(t *testing.T) {
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
		t.Run(
			tt.name, func(t *testing.T) {
				assert.Equal(t, tt.want, tt.score.String())
			},
		)
	}
}

func TestNewReviewScore(t *testing.T) {
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
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := flashcard.NewReviewScore(tt.args)

				assert.Equal(t, tt.want.err, err)
				assert.Equal(t, tt.want.score, got)
			},
		)
	}
}

/*
 Test Utilities
*/

var (
	oldestCard = flashcard.Card{
		Question:       "How do you delete a file?",
		Answer:         "import \"os\"\n\nos.Remove(path) error",
		EasinessFactor: 2.5,
		Interval:       0,
		Repetitions:    0,
		ReviewedAt:     time.Date(2021, 1, 2, 15, 0, 0, 0, time.UTC),
	}
	afterOldestCard  = oldestCard.ReviewedAt.Add(24 * time.Hour)
	beforeOldestCard = oldestCard.ReviewedAt.Add(-24 * time.Hour)
)
