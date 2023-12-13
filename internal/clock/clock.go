package clock

import (
	"time"
)

// Clock abstracts the time module to make it easier to test the system.
type Clock interface {
	Now() time.Time
}

// New creates a RealClock instance.
func New() RealClock {
	return RealClock{}
}

// RealClock uses the time module underneath and should be used the production code.
type RealClock struct{}

// Now returns the current local time.
func (RealClock) Now() time.Time {
	return time.Now()
}
