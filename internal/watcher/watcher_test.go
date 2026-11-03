package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewWatchesCreatedFile(t *testing.T) {
	dir := t.TempDir()

	w, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer w.Close()

	target := filepath.Join(dir, "hello.txt")
	if err := os.WriteFile(target, []byte("hi"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	select {
	case ev := <-w.Events():
		if ev.Name != target {
			t.Fatalf("got event for %s, want %s", ev.Name, target)
		}
	case err := <-w.Errors():
		t.Fatalf("unexpected watcher error: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for create event")
	}
}

func TestAddRecursiveWatchesSubdirectories(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	w, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer w.Close()

	target := filepath.Join(sub, "nested.txt")
	if err := os.WriteFile(target, []byte("hi"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	select {
	case ev := <-w.Events():
		if ev.Name != target {
			t.Fatalf("got event for %s, want %s", ev.Name, target)
		}
	case err := <-w.Errors():
		t.Fatalf("unexpected watcher error: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for create event in subdirectory")
	}
}
