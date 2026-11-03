// Package watcher wraps fsnotify with recursive directory watching so the
// core watch loop can be exercised independently of main().
package watcher

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"

	"github.com/rnikhori/go-file-watcher/internal/ignore"
)

// Options configures a Watcher.
type Options struct {
	// Root is the directory to watch recursively.
	Root string
	// Ignore, if non-nil, excludes matching paths from being watched and
	// reported.
	Ignore *ignore.Matcher
}

// Watcher recursively watches a root directory for filesystem events.
type Watcher struct {
	fsw       *fsnotify.Watcher
	root      string
	ignored   *ignore.Matcher
	closeOnce sync.Once
	closeErr  error
}

// New creates a Watcher rooted at opts.Root and adds watches for the root
// and all of its subdirectories, skipping any that match opts.Ignore.
func New(opts Options) (*Watcher, error) {
	absRoot, err := filepath.Abs(opts.Root)
	if err != nil {
		return nil, fmt.Errorf("resolving path: %w", err)
	}

	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("creating watcher: %w", err)
	}

	w := &Watcher{fsw: fsw, root: absRoot, ignored: opts.Ignore}
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

// Close releases the underlying fsnotify watcher. It is safe to call more
// than once; only the first call actually closes the watcher, and every
// call returns the same result.
func (w *Watcher) Close() error {
	w.closeOnce.Do(func() {
		w.closeErr = w.fsw.Close()
	})
	return w.closeErr
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
			if path != w.root && w.IsIgnored(path) {
				return filepath.SkipDir
			}
			return w.fsw.Add(path)
		}
		return nil
	})
}

// IsIgnored reports whether path matches the configured ignore patterns,
// relative to the watcher's root.
func (w *Watcher) IsIgnored(path string) bool {
	if w.ignored == nil {
		return false
	}
	rel, err := filepath.Rel(w.root, path)
	if err != nil {
		rel = path
	}
	return w.ignored.Match(rel)
}

// AddRecursive adds watches for dir and every subdirectory beneath it. It is
// exported so callers can add watches for directories created after start.
func (w *Watcher) AddRecursive(dir string) error {
	return w.addRecursive(dir)
}

// Remove removes the watch on dir, if one exists. Removing a path that was
// never watched (or whose watch was already cleaned up, e.g. by the kernel
// when the directory itself was deleted) is not an error.
func (w *Watcher) Remove(dir string) error {
	err := w.fsw.Remove(dir)
	if errors.Is(err, fsnotify.ErrNonExistentWatch) {
		return nil
	}
	return err
}
