package flashcard

import "time"

type Clock interface {
	Now() time.Time
}

// NewClock creates a RealClock instance.
func NewClock() RealClock {
	return RealClock{}
}

type RealClock struct {
}

// Now returns the current local time.
func (RealClock) Now() time.Time {
	return time.Now()
}
