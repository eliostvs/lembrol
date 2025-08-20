package flashcard

import (
	"time"

	fsrs "github.com/open-spaced-repetition/go-fsrs/v3"
)

// MigrateCard ensures a card has proper FSRS initialization.
// This function initializes cards that don't have FSRS data yet.
func MigrateCard(card Card, now time.Time) Card {
	// If already has FSRS data, return as-is
	if IsMigrated(card) {
		return card
	}
	
	// Initialize as a new FSRS card
	fsrsCard := fsrs.NewCard()
	
	migrated := card
	migrated.Due = now
	migrated.Stability = fsrsCard.Stability     // 0.0 for new cards
	migrated.Difficulty = fsrsCard.Difficulty   // 0.0 for new cards  
	migrated.ElapsedDays = fsrsCard.ElapsedDays // 0 for new cards
	migrated.ScheduledDays = fsrsCard.ScheduledDays // 0 for new cards
	migrated.Reps = fsrsCard.Reps               // 0 for new cards
	migrated.Lapses = countLapses(card.Stats)
	migrated.State = fsrsCard.State             // New for new cards
	migrated.LastReview = card.LastReview
	
	return migrated
}


// countLapses counts the number of failed reviews (Again ratings) from stats.
func countLapses(stats []Stats) uint64 {
	var lapses uint64
	for _, stat := range stats {
		if stat.Score == ReviewScoreAgain {
			lapses++
		}
	}
	return lapses
}

// IsMigrated checks if a card has been migrated to FSRS.
func IsMigrated(card Card) bool {
	// Card is considered migrated if it has any FSRS state set (even if values are zero)
	// A completely uninitialized card will have State = 0 (which is fsrs.New), 
	// but an uninitialized time.Time for LastReview
	return !card.LastReview.IsZero()
}

// MigrateDeck migrates all cards in a deck from SM-2 to FSRS.
func MigrateDeck(deck Deck, now time.Time) Deck {
	migratedCards := make([]Card, len(deck.Cards))
	
	for i, card := range deck.Cards {
		migratedCards[i] = MigrateCard(card, now)
	}
	
	deck.Cards = migratedCards
	return deck
}

