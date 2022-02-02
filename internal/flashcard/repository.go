package flashcard

// NewRepository creates a new Repository instance
// initializing the Deck and Stats repositories.
func NewRepository(location string, clock Clock) (*Repository, error) {
	deckRepository, err := NewDeckRepository(location, clock)
	if err != nil {
		return nil, err
	}
	return &Repository{Deck: deckRepository, Stats: NewStatsRepository(location)}, nil
}

// Repository wraps deck and stats repositories.
type Repository struct {
	Deck  *DeckRepository
	Stats *StatsRepository
}
