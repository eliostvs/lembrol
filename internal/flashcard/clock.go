package flashcard

import "time"

// Clock abstracts the time module to make it easier to test the system.
type Clock interface {
	Now() time.Time
}

// NewClock creates a RealClock instance.
func NewClock() RealClock {
	return RealClock{}
}

// RealClock uses the time module underneath and should use the production code.
type RealClock struct {
}

// Now returns the current local time.
func (RealClock) Now() time.Time {
	return time.Now()
}
