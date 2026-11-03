package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rnikhori/go-file-watcher/internal/ignore"
)

func TestIgnoredDirectoryIsNotWatched(t *testing.T) {
	dir := t.TempDir()
	ignoredDir := filepath.Join(dir, "vendor")
	if err := os.Mkdir(ignoredDir, 0o755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	w, err := New(Options{Root: dir, Ignore: ignore.New([]string{"vendor"})})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer w.Close()

	if err := os.WriteFile(filepath.Join(ignoredDir, "pkg.go"), []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	controlFile := filepath.Join(dir, "visible.txt")
	if err := os.WriteFile(controlFile, []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	select {
	case ev := <-w.Events():
		if ev.Name != controlFile {
			t.Fatalf("expected event only for %s, got event for %s (vendor directory should not be watched)", controlFile, ev.Name)
		}
	case err := <-w.Errors():
		t.Fatalf("unexpected watcher error: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for control file event")
	}
}

func TestIsIgnored(t *testing.T) {
	dir := t.TempDir()
	w, err := New(Options{Root: dir, Ignore: ignore.New([]string{"*.log"})})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer w.Close()

	if !w.IsIgnored(filepath.Join(dir, "debug.log")) {
		t.Error("expected debug.log to be ignored")
	}
	if w.IsIgnored(filepath.Join(dir, "main.go")) {
		t.Error("expected main.go to not be ignored")
	}
}
