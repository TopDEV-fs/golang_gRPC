// Package clock provides a simple time-source abstraction so that domain and
// application-layer code can be tested with a deterministic clock without
// depending on the system wall clock.
package clock

import "time"

// Clock is a source of the current wall-clock time.
type Clock interface {
	// Now returns the current UTC instant.
	Now() time.Time
}

// RealClock is the production implementation that delegates to time.Now().
type RealClock struct{}

// NewRealClock returns a RealClock ready for use.
func NewRealClock() RealClock {
	return RealClock{}
}

// Now returns the current time in UTC.
func (RealClock) Now() time.Time {
	return time.Now().UTC()
}
