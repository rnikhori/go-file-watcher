package app

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/rnikhori/go-file-watcher/internal/extfilter"
	"github.com/rnikhori/go-file-watcher/internal/ignore"
	"github.com/rnikhori/go-file-watcher/internal/output"
	"github.com/rnikhori/go-file-watcher/internal/watcher"
)

func newTestRunner(t *testing.T, root string, ignorePatterns []string, exts []string) (*Runner, *bytes.Buffer) {
	t.Helper()
	w, err := watcher.New(watcher.Options{Root: root, Ignore: ignore.New(ignorePatterns)})
	if err != nil {
		t.Fatalf("watcher.New: %v", err)
	}
	t.Cleanup(func() { w.Close() })

	var buf bytes.Buffer
	return &Runner{
		Watcher:   w,
		Ext:       extfilter.New(exts),
		Formatter: output.NewText(&buf),
	}, &buf
}

func TestHandleEventWritesSurvivingEvent(t *testing.T) {
	dir := t.TempDir()
	r, buf := newTestRunner(t, dir, nil, nil)

	target := filepath.Join(dir, "foo.go")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	wrote := r.HandleEvent(fsnotify.Event{Name: target, Op: fsnotify.Create}, time.Now())
	if !wrote {
		t.Fatal("expected event to be written")
	}
	if !bytes.Contains(buf.Bytes(), []byte("foo.go")) {
		t.Errorf("output missing event: %q", buf.String())
	}
}

func TestHandleEventSuppressesIgnoredPath(t *testing.T) {
	dir := t.TempDir()
	r, buf := newTestRunner(t, dir, []string{"*.log"}, nil)

	target := filepath.Join(dir, "debug.log")
	wrote := r.HandleEvent(fsnotify.Event{Name: target, Op: fsnotify.Write}, time.Now())
	if wrote {
		t.Fatal("expected ignored path to be suppressed")
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output, got %q", buf.String())
	}
}

func TestHandleEventSuppressesNonMatchingExtension(t *testing.T) {
	dir := t.TempDir()
	r, buf := newTestRunner(t, dir, nil, []string{".go"})

	target := filepath.Join(dir, "notes.txt")
	wrote := r.HandleEvent(fsnotify.Event{Name: target, Op: fsnotify.Write}, time.Now())
	if wrote {
		t.Fatal("expected non-matching extension to be suppressed")
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output, got %q", buf.String())
	}
}

func TestHandleEventAddsWatchForNewSubdirectory(t *testing.T) {
	dir := t.TempDir()
	r, _ := newTestRunner(t, dir, nil, nil)

	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	if wrote := r.HandleEvent(fsnotify.Event{Name: sub, Op: fsnotify.Create}, time.Now()); !wrote {
		t.Fatal("expected directory create event to be written")
	}

	nested := filepath.Join(sub, "nested.txt")
	if err := os.WriteFile(nested, []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	deadline := time.After(5 * time.Second)
	for {
		select {
		case ev := <-r.Watcher.Events():
			if ev.Name == nested {
				return
			}
		case <-deadline:
			t.Fatal("timed out waiting for event in newly added subdirectory")
		}
	}
}
