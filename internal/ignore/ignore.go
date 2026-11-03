// Package ignore implements gitignore-like glob matching used to exclude
// paths from being watched or reported.
package ignore

import (
	"path/filepath"
	"strings"
)

// Matcher holds a set of glob patterns and reports whether a given path
// matches any of them.
type Matcher struct {
	patterns []string
}

// New builds a Matcher from patterns. Blank entries and lines starting with
// '#' are treated as comments and skipped, mirroring .gitignore conventions.
func New(patterns []string) *Matcher {
	m := &Matcher{}
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" || strings.HasPrefix(p, "#") {
			continue
		}
		m.patterns = append(m.patterns, filepath.ToSlash(p))
	}
	return m
}

// Match reports whether relPath, or any path segment within it, matches one
// of the configured patterns. relPath is expected to be relative to the
// watched root.
func (m *Matcher) Match(relPath string) bool {
	if m == nil || len(m.patterns) == 0 {
		return false
	}

	relPath = filepath.ToSlash(relPath)
	base := filepath.Base(relPath)

	for _, p := range m.patterns {
		if matched, _ := filepath.Match(p, relPath); matched {
			return true
		}
		if matched, _ := filepath.Match(p, base); matched {
			return true
		}
		for _, seg := range strings.Split(relPath, "/") {
			if seg == "" {
				continue
			}
			if matched, _ := filepath.Match(p, seg); matched {
				return true
			}
		}
	}
	return false
}

// Patterns returns the normalized patterns held by the matcher.
func (m *Matcher) Patterns() []string {
	if m == nil {
		return nil
	}
	return append([]string(nil), m.patterns...)
}
