// Package app wires the watcher, filters, and output formatter together
// into a single run loop, kept separate from main() so it can be unit
// tested directly.
package app

import (
	"log"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/rnikhori/go-file-watcher/internal/debounce"
	"github.com/rnikhori/go-file-watcher/internal/extfilter"
	"github.com/rnikhori/go-file-watcher/internal/output"
	"github.com/rnikhori/go-file-watcher/internal/watcher"
)

// Runner drives the watch loop: it reads events from a Watcher, applies
// filters, and writes surviving events through a Formatter.
type Runner struct {
	Watcher   *watcher.Watcher
	Ext       *extfilter.Filter
	Formatter output.Formatter
	// Debounce, if non-nil, coalesces rapid repeated events for the same
	// path within its window into a single reported event.
	Debounce *debounce.Coalescer
}

// HandleEvent applies ignore/extension filtering to a single fsnotify event,
// keeps the watcher's directory tree in sync for Create events, and writes
// the event through the Formatter if it was not filtered out. It reports
// whether the event was written.
func (r *Runner) HandleEvent(event fsnotify.Event, now time.Time) bool {
	if r.Watcher.IsIgnored(event.Name) {
		return false
	}

	isDir := false
	switch {
	case event.Op&fsnotify.Create != 0:
		if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
			isDir = true
			if err := r.Watcher.AddRecursive(event.Name); err != nil {
				log.Printf("watching new dir %s: %v", event.Name, err)
			}
		}
	case event.Op&(fsnotify.Remove|fsnotify.Rename) != 0:
		if err := r.Watcher.Remove(event.Name); err != nil {
			log.Printf("unwatching removed dir %s: %v", event.Name, err)
		}
	}

	if !isDir && !r.Ext.Match(event.Name) {
		return false
	}

	if r.Debounce != nil && !r.Debounce.Allow(event.Name, now) {
		return false
	}

	if err := r.Formatter.Write(event, now); err != nil {
		log.Printf("writing event: %v", err)
	}
	return true
}

// Run drives the watch loop until the watcher's channels close or stop is
// signaled. It returns when the loop exits.
func (r *Runner) Run(stop <-chan os.Signal) {
	for {
		select {
		case event, ok := <-r.Watcher.Events():
			if !ok {
				return
			}
			r.HandleEvent(event, time.Now())
		case err, ok := <-r.Watcher.Errors():
			if !ok {
				return
			}
			log.Printf("watcher error: %v", err)
		case <-stop:
			return
		}
	}
}
