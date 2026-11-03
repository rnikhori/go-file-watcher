// Package debounce coalesces rapid repeated events for the same key into a
// single reported event within a configurable time window.
package debounce

import "time"

// Coalescer suppresses events for a key that occur within window of the
// previously allowed event for that key, so a burst of rapid events for one
// path is reported once rather than once per underlying event.
type Coalescer struct {
	window time.Duration
	last   map[string]time.Time
}

// New creates a Coalescer with the given debounce window. A non-positive
// window disables debouncing: every call to Allow returns true.
func New(window time.Duration) *Coalescer {
	return &Coalescer{window: window, last: make(map[string]time.Time)}
}

// Allow reports whether an event for key at time now should be reported. It
// returns false if an event for the same key was already allowed within the
// debounce window preceding now; otherwise it records now as the latest
// allowed time for key and returns true.
func (c *Coalescer) Allow(key string, now time.Time) bool {
	if c.window <= 0 {
		return true
	}
	if last, ok := c.last[key]; ok && now.Sub(last) < c.window {
		return false
	}
	c.last[key] = now
	return true
}
