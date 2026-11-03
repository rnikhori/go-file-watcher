// Package watcher wraps fsnotify with recursive directory watching so the
// core watch loop can be exercised independently of main().
package watcher

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// Watcher recursively watches a root directory for filesystem events.
type Watcher struct {
	fsw  *fsnotify.Watcher
	root string
}

// New creates a Watcher rooted at root and adds watches for root and all of
// its subdirectories.
func New(root string) (*Watcher, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolving path: %w", err)
	}

	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("creating watcher: %w", err)
	}

	w := &Watcher{fsw: fsw, root: absRoot}
	if err := w.addRecursive(absRoot); err != nil {
		fsw.Close()
		return nil, fmt.Errorf("watching %s: %w", absRoot, err)
	}
	return w, nil
}

// Root returns the absolute path the watcher was created with.
func (w *Watcher) Root() string {
	return w.root
}

// Close releases the underlying fsnotify watcher.
func (w *Watcher) Close() error {
	return w.fsw.Close()
}

// Events exposes the underlying fsnotify event channel.
func (w *Watcher) Events() <-chan fsnotify.Event {
	return w.fsw.Events
}

// Errors exposes the underlying fsnotify error channel.
func (w *Watcher) Errors() <-chan error {
	return w.fsw.Errors
}

func (w *Watcher) addRecursive(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return w.fsw.Add(path)
		}
		return nil
	})
}

// AddRecursive adds watches for dir and every subdirectory beneath it. It is
// exported so callers can add watches for directories created after start.
func (w *Watcher) AddRecursive(dir string) error {
	return w.addRecursive(dir)
}

// Remove removes the watch on dir, if one exists.
func (w *Watcher) Remove(dir string) error {
	return w.fsw.Remove(dir)
}
