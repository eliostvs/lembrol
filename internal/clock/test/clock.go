package test

import "time"

func New(t time.Time) Clock {
	return Clock{t}
}

type Clock struct {
	time.Time
}

func (c Clock) Now() time.Time {
	return c.Time
}
