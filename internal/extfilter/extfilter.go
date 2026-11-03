// Package extfilter restricts reported events to files with a configured
// set of extensions.
package extfilter

import (
	"path/filepath"
	"strings"
)

// Filter matches file paths against a set of extensions.
type Filter struct {
	exts map[string]bool
}

// New builds a Filter from extensions such as ".go" or "go" (the leading
// dot is optional and matching is case-insensitive). An empty list means
// everything matches.
func New(exts []string) *Filter {
	f := &Filter{exts: make(map[string]bool)}
	for _, e := range exts {
		e = strings.ToLower(strings.TrimSpace(e))
		if e == "" {
			continue
		}
		if !strings.HasPrefix(e, ".") {
			e = "." + e
		}
		f.exts[e] = true
	}
	return f
}

// Match reports whether path's extension is in the configured set. A Filter
// with no extensions configured matches every path, including those with no
// extension at all.
func (f *Filter) Match(path string) bool {
	if f == nil || len(f.exts) == 0 {
		return true
	}
	return f.exts[strings.ToLower(filepath.Ext(path))]
}
