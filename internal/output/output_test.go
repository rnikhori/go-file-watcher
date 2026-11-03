package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestTextFormatterWrite(t *testing.T) {
	var buf bytes.Buffer
	f := NewText(&buf)
	when := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	event := fsnotify.Event{Name: "/tmp/foo.go", Op: fsnotify.Write}

	if err := f.Write(event, when); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "2024-01-02 03:04:05") {
		t.Errorf("output missing timestamp: %q", got)
	}
	if !strings.Contains(got, "/tmp/foo.go") {
		t.Errorf("output missing path: %q", got)
	}
	if !strings.Contains(got, "WRITE") {
		t.Errorf("output missing op: %q", got)
	}
}

func TestJSONFormatterWrite(t *testing.T) {
	var buf bytes.Buffer
	f := NewJSON(&buf)
	when := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	event := fsnotify.Event{Name: "/tmp/foo.go", Op: fsnotify.Create}

	if err := f.Write(event, when); err != nil {
		t.Fatalf("Write: %v", err)
	}

	var decoded Event
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("Unmarshal: %v (output: %q)", err, buf.String())
	}
	if decoded.Path != "/tmp/foo.go" {
		t.Errorf("Path = %q, want %q", decoded.Path, "/tmp/foo.go")
	}
	if decoded.Op != "CREATE" {
		t.Errorf("Op = %q, want %q", decoded.Op, "CREATE")
	}
	if decoded.Time != when.Format(time.RFC3339) {
		t.Errorf("Time = %q, want %q", decoded.Time, when.Format(time.RFC3339))
	}
}

func TestJSONFormatterOneObjectPerLine(t *testing.T) {
	var buf bytes.Buffer
	f := NewJSON(&buf)
	when := time.Now()

	for i := 0; i < 3; i++ {
		if err := f.Write(fsnotify.Event{Name: "a", Op: fsnotify.Write}, when); err != nil {
			t.Fatalf("Write: %v", err)
		}
	}

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3: %q", len(lines), buf.String())
	}
	for _, line := range lines {
		var decoded Event
		if err := json.Unmarshal([]byte(line), &decoded); err != nil {
			t.Errorf("line %q is not valid JSON: %v", line, err)
		}
	}
}
