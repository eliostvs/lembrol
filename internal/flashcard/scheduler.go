package flashcard

import (
	"time"

	fsrs "github.com/open-spaced-repetition/go-fsrs/v3"
)

// Scheduler wraps the FSRS algorithm for card scheduling.
type Scheduler struct {
	fsrs *fsrs.FSRS
}

// NewScheduler creates a new FSRS-based scheduler with the given parameters.
func NewScheduler(params fsrs.Parameters) *Scheduler {
	return &Scheduler{
		fsrs: fsrs.NewFSRS(params),
	}
}

// DefaultScheduler creates a new scheduler with default FSRS parameters.
func DefaultScheduler() *Scheduler {
	return NewScheduler(fsrs.DefaultParam())
}

// ScheduleCard schedules a card review with the given rating.
func (s *Scheduler) ScheduleCard(card Card, now time.Time, rating fsrs.Rating) (Card, Stats) {
	// Convert lembrol Card to FSRS Card
	fsrsCard := s.cardToFSRS(card)

	// Get scheduling info from FSRS
	info := s.fsrs.Next(fsrsCard, now, rating)

	// Convert back to lembrol Card
	updatedCard := s.fsrsToCard(info.Card, card)

	// Create stats record
	stats := NewFSRSStats(now, rating, card, updatedCard)

	return updatedCard, stats
}

// GetReviewOptions returns all possible review outcomes for a card.
func (s *Scheduler) GetReviewOptions(card Card, now time.Time) map[fsrs.Rating]Card {
	fsrsCard := s.cardToFSRS(card)
	outcomes := s.fsrs.Repeat(fsrsCard, now)

	options := make(map[fsrs.Rating]Card)
	for rating, log := range outcomes {
		options[rating] = s.fsrsToCard(log.Card, card)
	}

	return options
}

// GetRetrievability returns the current retrievability of a card.
func (s *Scheduler) GetRetrievability(card Card, now time.Time) float64 {
	fsrsCard := s.cardToFSRS(card)
	return s.fsrs.GetRetrievability(fsrsCard, now)
}

// cardToFSRS converts a lembrol Card to an FSRS Card.
func (s *Scheduler) cardToFSRS(card Card) fsrs.Card {
	return fsrs.Card{
		Due:           card.Due,
		Stability:     card.Stability,
		Difficulty:    card.Difficulty,
		ElapsedDays:   card.ElapsedDays,
		ScheduledDays: card.ScheduledDays,
		Reps:          card.Reps,
		Lapses:        card.Lapses,
		State:         card.State,
		LastReview:    card.LastReview,
	}
}

// fsrsToCard converts an FSRS Card back to a lembrol Card, preserving metadata.
func (s *Scheduler) fsrsToCard(fsrsCard fsrs.Card, original Card) Card {
	return Card{
		// Preserve original metadata
		ID:       original.ID,
		Question: original.Question,
		Answer:   original.Answer,
		Stats:    original.Stats,

		// Updated LastReview timestamp is handled by FSRS fields below

		// FSRS fields
		Due:           fsrsCard.Due,
		Stability:     fsrsCard.Stability,
		Difficulty:    fsrsCard.Difficulty,
		ElapsedDays:   fsrsCard.ElapsedDays,
		ScheduledDays: fsrsCard.ScheduledDays,
		Reps:          fsrsCard.Reps,
		Lapses:        fsrsCard.Lapses,
		State:         fsrsCard.State,
		LastReview:    fsrsCard.LastReview,

	}
}

// ReviewScoreToFSRSRating converts the current ReviewScore to FSRS Rating.
func ReviewScoreToFSRSRating(score ReviewScore) fsrs.Rating {
	switch score {
	case ReviewScoreAgain:
		return fsrs.Again
	case ReviewScoreHard:
		return fsrs.Hard
	case ReviewScoreNormal:
		return fsrs.Good
	case ReviewScoreEasy, ReviewScoreSuperEasy:
		return fsrs.Easy
	default:
		return fsrs.Good
	}
}

// FSRSRatingToReviewScore converts FSRS Rating back to ReviewScore for compatibility.
func FSRSRatingToReviewScore(rating fsrs.Rating) ReviewScore {
	switch rating {
	case fsrs.Again:
		return ReviewScoreAgain
	case fsrs.Hard:
		return ReviewScoreHard
	case fsrs.Good:
		return ReviewScoreNormal
	case fsrs.Easy:
		return ReviewScoreEasy
	default:
		return ReviewScoreNormal
	}
}
