package app

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/rnikhori/go-file-watcher/internal/debounce"
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

func TestHandleEventDebounceCoalescesBurst(t *testing.T) {
	dir := t.TempDir()
	r, buf := newTestRunner(t, dir, nil, nil)
	r.Debounce = debounce.New(100 * time.Millisecond)

	target := filepath.Join(dir, "hot.txt")
	base := time.Unix(0, 0)

	wroteCount := 0
	for i := 0; i < 5; i++ {
		if r.HandleEvent(fsnotify.Event{Name: target, Op: fsnotify.Write}, base.Add(time.Duration(i)*10*time.Millisecond)) {
			wroteCount++
		}
	}
	if wroteCount != 1 {
		t.Fatalf("expected a burst within the debounce window to produce 1 write, got %d", wroteCount)
	}

	if !r.HandleEvent(fsnotify.Event{Name: target, Op: fsnotify.Write}, base.Add(200*time.Millisecond)) {
		t.Fatal("expected event after the debounce window to be written")
	}

	lines := bytes.Count(buf.Bytes(), []byte("\n"))
	if lines != 2 {
		t.Fatalf("expected 2 reported lines, got %d: %q", lines, buf.String())
	}
}

func TestHandleEventRemovesWatchForDeletedSubdirectory(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	r, _ := newTestRunner(t, dir, nil, nil)

	if err := os.RemoveAll(sub); err != nil {
		t.Fatalf("RemoveAll: %v", err)
	}

	deadline := time.After(5 * time.Second)
	var sawRemove bool
	for !sawRemove {
		select {
		case ev := <-r.Watcher.Events():
			if ev.Name == sub && ev.Op&fsnotify.Remove != 0 {
				sawRemove = true
				if !r.HandleEvent(ev, time.Now()) {
					t.Fatal("expected remove event for watched dir to be written")
				}
			}
		case err := <-r.Watcher.Errors():
			t.Fatalf("unexpected watcher error after removing a watched subdirectory: %v", err)
		case <-deadline:
			t.Fatal("timed out waiting for remove event")
		}
	}

	if err := r.Watcher.Remove(sub); err != nil {
		t.Fatalf("Remove on an already-cleaned-up watch should be a no-op, got: %v", err)
	}

	control := filepath.Join(dir, "still-watching.txt")
	if err := os.WriteFile(control, []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	select {
	case ev := <-r.Watcher.Events():
		if ev.Name != control {
			t.Fatalf("got event for %s, want %s", ev.Name, control)
		}
	case err := <-r.Watcher.Errors():
		t.Fatalf("unexpected watcher error: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for control event; watcher on root may be broken")
	}
}

func TestRunOnceExitsAfterFirstMatchingEvent(t *testing.T) {
	dir := t.TempDir()
	r, buf := newTestRunner(t, dir, nil, []string{".go"})
	r.Once = true

	done := make(chan struct{})
	go func() {
		r.Run(make(chan os.Signal))
		close(done)
	}()

	if err := os.WriteFile(filepath.Join(dir, "skip.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	target := filepath.Join(dir, "match.go")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after the first matching event")
	}

	if !bytes.Contains(buf.Bytes(), []byte("match.go")) {
		t.Errorf("expected output to contain the matching event, got %q", buf.String())
	}
	if bytes.Contains(buf.Bytes(), []byte("skip.txt")) {
		t.Errorf("expected non-matching event to be filtered out, got %q", buf.String())
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
