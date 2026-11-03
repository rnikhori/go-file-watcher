package watcher

import (
	"runtime"
	"testing"
	"time"
)

func TestCloseIsIdempotent(t *testing.T) {
	w, err := New(Options{Root: t.TempDir()})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("second Close should also succeed, got: %v", err)
	}
}

func TestRepeatedCreateCloseDoesNotLeakGoroutines(t *testing.T) {
	before := runtime.NumGoroutine()

	for i := 0; i < 20; i++ {
		w, err := New(Options{Root: t.TempDir()})
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		if err := w.Close(); err != nil {
			t.Fatalf("Close: %v", err)
		}
	}

	// fsnotify's internal goroutines stop asynchronously after Close, so
	// give them a moment before comparing counts.
	deadline := time.Now().Add(2 * time.Second)
	var after int
	for {
		after = runtime.NumGoroutine()
		if after <= before+2 || time.Now().After(deadline) {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if after > before+2 {
		t.Errorf("goroutine count grew from %d to %d after 20 create/close cycles", before, after)
	}
}
