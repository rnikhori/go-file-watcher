// Package summary tracks counts of filesystem events by operation type for
// end-of-run reporting.
package summary

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fsnotify/fsnotify"
)

// Counter accumulates event counts keyed by fsnotify.Op string
// representation (e.g. "CREATE", "WRITE").
type Counter struct {
	counts map[string]int
	total  int
}

// NewCounter creates an empty Counter.
func NewCounter() *Counter {
	return &Counter{counts: make(map[string]int)}
}

// Add records one occurrence of op.
func (c *Counter) Add(op fsnotify.Op) {
	c.counts[op.String()]++
	c.total++
}

// Total returns the number of events recorded across all operation types.
func (c *Counter) Total() int {
	return c.total
}

// Counts returns a copy of the per-operation counts.
func (c *Counter) Counts() map[string]int {
	out := make(map[string]int, len(c.counts))
	for k, v := range c.counts {
		out[k] = v
	}
	return out
}

// String renders the counts as a single line, operations sorted
// alphabetically for stable output, e.g. "total=5 CREATE=2 WRITE=3".
func (c *Counter) String() string {
	keys := make([]string, 0, len(c.counts))
	for k := range c.counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys)+1)
	parts = append(parts, fmt.Sprintf("total=%d", c.total))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", k, c.counts[k]))
	}
	return strings.Join(parts, " ")
}
