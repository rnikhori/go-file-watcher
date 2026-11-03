package ignore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFileMissingFileReturnsNoError(t *testing.T) {
	patterns, err := LoadFile(filepath.Join(t.TempDir(), "does-not-exist"))
	if err != nil {
		t.Fatalf("LoadFile on missing file returned error: %v", err)
	}
	if len(patterns) != 0 {
		t.Fatalf("expected no patterns, got %v", patterns)
	}
}

func TestLoadFileReadsLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".watchignore")
	content := "*.log\n# comment\n\nnode_modules\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	patterns, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	m := New(patterns)
	if !m.Match("debug.log") {
		t.Error("expected debug.log to be ignored")
	}
	if !m.Match("node_modules/pkg/index.js") {
		t.Error("expected node_modules subpath to be ignored")
	}
	if m.Match("main.go") {
		t.Error("expected main.go to not be ignored")
	}
}
