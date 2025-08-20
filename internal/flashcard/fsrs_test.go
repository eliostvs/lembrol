package flashcard_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eliostvs/lembrol/internal/flashcard"
	fsrs "github.com/open-spaced-repetition/go-fsrs/v3"
)

func TestNewCard_FSRS(t *testing.T) {
	now := time.Now()
	card := flashcard.NewCard("What is 2+2?", "4", now)

	assert.Equal(t, "What is 2+2?", card.Question)
	assert.Equal(t, "4", card.Answer)
	assert.Equal(t, now, card.LastReview)
	assert.Equal(t, now, card.Due)
	assert.Equal(t, fsrs.New, card.State)
	assert.Zero(t, card.Difficulty) // New cards have zero difficulty until first review
	assert.Zero(t, card.Stability) // New cards have zero stability until first review
	assert.Equal(t, uint64(0), card.Reps)
	assert.Equal(t, uint64(0), card.Lapses)
}

func TestScheduler_ScheduleCard(t *testing.T) {
	scheduler := flashcard.DefaultScheduler()
	now := time.Now()
	card := flashcard.NewCard("Test", "Answer", now)

	tests := []struct {
		name   string
		rating fsrs.Rating
	}{
		{"Again", fsrs.Again},
		{"Hard", fsrs.Hard},
		{"Good", fsrs.Good},
		{"Easy", fsrs.Easy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedCard, stats := scheduler.ScheduleCard(card, now, tt.rating)

			// Algorithm and Timestamp fields removed - no longer testing them
			assert.Equal(t, tt.rating, stats.Rating)
			assert.NotZero(t, updatedCard.Stability)
			assert.NotZero(t, updatedCard.Difficulty)
			assert.True(t, updatedCard.Due.After(now) || updatedCard.Due.Equal(now))
		})
	}
}

func TestCard_AdvanceWithFSRS(t *testing.T) {
	now := time.Now()
	card := flashcard.NewCard("Test", "Answer", now)
	scheduler := flashcard.DefaultScheduler()

	// Test progression: Good -> Good -> Good
	ratings := []flashcard.ReviewScore{
		flashcard.ReviewScoreNormal, // Good
		flashcard.ReviewScoreNormal, // Good
		flashcard.ReviewScoreNormal, // Good
	}

	currentCard := card
	for i, score := range ratings {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			updatedCard := currentCard.AdvanceWithScheduler(now, score, scheduler)
			
			assert.Greater(t, updatedCard.Reps, currentCard.Reps)
			assert.True(t, updatedCard.Due.After(now))
			assert.Len(t, updatedCard.Stats, len(currentCard.Stats)+1)
			
			// Verify stats
			lastStat := updatedCard.Stats[len(updatedCard.Stats)-1]
			// Algorithm field removed - no longer testing it
			assert.Equal(t, score, lastStat.Score)
			
			currentCard = updatedCard
			now = now.Add(24 * time.Hour) // Move forward a day
		})
	}
}

func TestCard_AdvanceWithAgainRating(t *testing.T) {
	now := time.Now()
	card := flashcard.NewCard("Test", "Answer", now)
	scheduler := flashcard.DefaultScheduler()

	// Progress card through learning stages to Review state
	card = card.AdvanceWithScheduler(now, flashcard.ReviewScoreNormal, scheduler)
	card = card.AdvanceWithScheduler(now.Add(24*time.Hour), flashcard.ReviewScoreNormal, scheduler)
	card = card.AdvanceWithScheduler(now.Add(48*time.Hour), flashcard.ReviewScoreNormal, scheduler)
	
	// Now card should be in Review state, lapses will increment on "Again"
	initialLapses := card.Lapses

	// Then fail it with Again
	failedCard := card.AdvanceWithScheduler(now.Add(72*time.Hour), flashcard.ReviewScoreAgain, scheduler)

	assert.Greater(t, failedCard.Lapses, initialLapses)
	assert.Greater(t, len(failedCard.Stats), 3) // Should have at least 4 stats entries
	
	// Verify the failure was recorded
	lastStat := failedCard.Stats[len(failedCard.Stats)-1]
	// Algorithm field removed - no longer testing it
	assert.Equal(t, flashcard.ReviewScoreAgain, lastStat.Score)
	assert.Equal(t, fsrs.Again, lastStat.Rating)
}

func TestMigrateCard(t *testing.T) {
	now := time.Now()
	
	// Create a card without FSRS data
	emptyCard := flashcard.Card{
		ID:         "test-id",
		Question:   "What is 2+2?",
		Answer:     "4",
		LastReview: now.Add(-24 * time.Hour),
		Stats: []flashcard.Stats{
			{
				Score:      flashcard.ReviewScoreNormal,
				LastReview: now.Add(-48 * time.Hour),
			},
		},
	}

	migratedCard := flashcard.MigrateCard(emptyCard, now)

	assert.Equal(t, emptyCard.ID, migratedCard.ID)
	assert.Equal(t, emptyCard.Question, migratedCard.Question)
	assert.Equal(t, emptyCard.Answer, migratedCard.Answer)
	assert.Equal(t, uint64(0), migratedCard.Reps)
	assert.Zero(t, migratedCard.Difficulty) // Migrated cards start as new FSRS cards with zero values
	assert.Zero(t, migratedCard.Stability) // Migrated cards start as new FSRS cards with zero values
	assert.Equal(t, fsrs.New, migratedCard.State)
	assert.True(t, flashcard.IsMigrated(migratedCard))
}

func TestIsMigrated(t *testing.T) {
	now := time.Now()
	
	// New FSRS card
	fsrsCard := flashcard.NewCard("Test", "Answer", now)
	assert.True(t, flashcard.IsMigrated(fsrsCard))
	
	// Empty card without FSRS data (zero LastReview time means not migrated)
	emptyCard := flashcard.Card{
		ID:       "test",
		Question: "Test",
		Answer:   "Answer",
		// LastReview is zero time, indicating not migrated
	}
	assert.False(t, flashcard.IsMigrated(emptyCard))
}

func TestReviewScoreToFSRSRating(t *testing.T) {
	tests := []struct {
		score    flashcard.ReviewScore
		expected fsrs.Rating
	}{
		{flashcard.ReviewScoreAgain, fsrs.Again},
		{flashcard.ReviewScoreHard, fsrs.Hard},
		{flashcard.ReviewScoreNormal, fsrs.Good},
		{flashcard.ReviewScoreEasy, fsrs.Easy},
		{flashcard.ReviewScoreSuperEasy, fsrs.Easy}, // Maps to Easy
	}

	for _, tt := range tests {
		t.Run(tt.score.String(), func(t *testing.T) {
			result := flashcard.ReviewScoreToFSRSRating(tt.score)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFSRSRatingToReviewScore(t *testing.T) {
	tests := []struct {
		rating   fsrs.Rating
		expected flashcard.ReviewScore
	}{
		{fsrs.Again, flashcard.ReviewScoreAgain},
		{fsrs.Hard, flashcard.ReviewScoreHard},
		{fsrs.Good, flashcard.ReviewScoreNormal},
		{fsrs.Easy, flashcard.ReviewScoreEasy},
	}

	for _, tt := range tests {
		t.Run(string(rune('0'+int(tt.rating))), func(t *testing.T) {
			result := flashcard.FSRSRatingToReviewScore(tt.rating)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCard_NextReviewAt_FSRS(t *testing.T) {
	now := time.Now()
	futureTime := now.Add(24 * time.Hour)
	
	// FSRS card with Due set
	fsrsCard := flashcard.Card{
		Due: futureTime,
	}
	assert.Equal(t, futureTime, fsrsCard.NextReviewAt())
	
	// Card without due date
	cardWithoutDue := flashcard.Card{
		LastReview: now,
	}
	assert.Equal(t, now, cardWithoutDue.NextReviewAt())
}

func TestCard_IsDue_FSRS(t *testing.T) {
	now := time.Now()
	futureTime := now.Add(24 * time.Hour)
	pastTime := now.Add(-24 * time.Hour)
	
	// FSRS card due in the future
	futureCard := flashcard.Card{Due: futureTime}
	assert.False(t, futureCard.IsDue(now))
	
	// FSRS card due in the past
	pastCard := flashcard.Card{Due: pastTime}
	assert.True(t, pastCard.IsDue(now))
	
	// FSRS card due now
	nowCard := flashcard.Card{Due: now}
	assert.True(t, nowCard.IsDue(now))
}

func TestNewFSRSStats(t *testing.T) {
	now := time.Now()
	previous := flashcard.Card{
		LastReview: now.Add(-time.Hour),
		Reps:       1,
	}
	updated := flashcard.Card{
		Stability:     2.5,
		Difficulty:    3.2,
		ElapsedDays:   1,
		ScheduledDays: 3,
		State:         fsrs.Review,
	}

	stats := flashcard.NewFSRSStats(now, fsrs.Good, previous, updated)

	// Algorithm and Timestamp fields removed - no longer testing them
	assert.Equal(t, flashcard.ReviewScoreNormal, stats.Score)
	assert.Equal(t, fsrs.Good, stats.Rating)
	assert.Equal(t, now, stats.LastReview) // Should use review timestamp, not previous review
	assert.Equal(t, updated.Stability, stats.Stability)
	assert.Equal(t, updated.Difficulty, stats.Difficulty)
	assert.Equal(t, updated.ElapsedDays, stats.ElapsedDays)
	assert.Equal(t, updated.ScheduledDays, stats.ScheduledDays)
	assert.Equal(t, updated.State, stats.State)
}

func TestConfig_ToFSRSParameters(t *testing.T) {
	config := flashcard.Config{
		RequestRetention: 0.9,
		MaximumInterval:  10000,
		EnableShortTerm:  true,
		EnableFuzz:       false,
		Weights:          nil, // Use defaults
	}

	params := config.ToFSRSParameters()

	assert.Equal(t, 0.9, params.RequestRetention)
	assert.Equal(t, 10000.0, params.MaximumInterval)
	assert.True(t, params.EnableShortTerm)
	assert.False(t, params.EnableFuzz)
}

func TestMigrateDeck(t *testing.T) {
	now := time.Now()
	
	cards := []flashcard.Card{
		{
			ID:         "card1",
			Question:   "Q1",
			Answer:     "A1",
			LastReview: now.Add(-24 * time.Hour),
		},
		{
			ID:         "card2", 
			Question:   "Q2",
			Answer:     "A2",
			LastReview: now.Add(-48 * time.Hour),
		},
	}
	
	deck := flashcard.Deck{
		Name:  "Test Deck",
		Cards: cards,
	}

	migratedDeck := flashcard.MigrateDeck(deck, now)

	require.Len(t, migratedDeck.Cards, 2)
	for _, card := range migratedDeck.Cards {
		assert.True(t, flashcard.IsMigrated(card))
		assert.Zero(t, card.Difficulty) // Migrated cards start as new FSRS cards with zero values
		assert.Zero(t, card.Stability) // Migrated cards start as new FSRS cards with zero values
	}
}