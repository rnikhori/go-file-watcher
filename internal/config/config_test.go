package config

import (
	"reflect"
	"testing"
	"time"
)

func TestParseDefaults(t *testing.T) {
	cfg, err := Parse(nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if cfg.Root != "." {
		t.Errorf("Root = %q, want %q", cfg.Root, ".")
	}
	if cfg.IgnoreFile != ".watchignore" {
		t.Errorf("IgnoreFile = %q, want %q", cfg.IgnoreFile, ".watchignore")
	}
	if len(cfg.IgnorePatterns) != 0 {
		t.Errorf("IgnorePatterns = %v, want empty", cfg.IgnorePatterns)
	}
}

func TestParsePositionalRoot(t *testing.T) {
	cfg, err := Parse([]string{"/tmp/some-dir"})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if cfg.Root != "/tmp/some-dir" {
		t.Errorf("Root = %q, want %q", cfg.Root, "/tmp/some-dir")
	}
}

func TestParseIgnoreFlag(t *testing.T) {
	cfg, err := Parse([]string{"--ignore", "*.log, node_modules ,build"})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	want := []string{"*.log", "node_modules", "build"}
	if !reflect.DeepEqual(cfg.IgnorePatterns, want) {
		t.Errorf("IgnorePatterns = %v, want %v", cfg.IgnorePatterns, want)
	}
}

func TestParseExtFlag(t *testing.T) {
	cfg, err := Parse([]string{"--ext", ".go, .txt ,md"})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	want := []string{".go", ".txt", "md"}
	if !reflect.DeepEqual(cfg.Ext, want) {
		t.Errorf("Ext = %v, want %v", cfg.Ext, want)
	}
}

func TestParseJSONFlag(t *testing.T) {
	cfg, err := Parse([]string{"--json"})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !cfg.JSON {
		t.Error("JSON = false, want true")
	}

	cfg, err = Parse(nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if cfg.JSON {
		t.Error("JSON = true by default, want false")
	}
}

func TestParseDebounceFlag(t *testing.T) {
	cfg, err := Parse([]string{"--debounce", "250ms"})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if cfg.Debounce != 250*time.Millisecond {
		t.Errorf("Debounce = %v, want %v", cfg.Debounce, 250*time.Millisecond)
	}

	cfg, err = Parse(nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if cfg.Debounce != 0 {
		t.Errorf("Debounce default = %v, want 0", cfg.Debounce)
	}
}

func TestParseOnceFlag(t *testing.T) {
	cfg, err := Parse([]string{"--once"})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !cfg.Once {
		t.Error("Once = false, want true")
	}
}

func TestParseSummaryFlag(t *testing.T) {
	cfg, err := Parse([]string{"--summary"})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !cfg.Summary {
		t.Error("Summary = false, want true")
	}
}

func TestParseInvalidFlag(t *testing.T) {
	if _, err := Parse([]string{"--not-a-real-flag"}); err == nil {
		t.Fatal("expected error for unknown flag")
	}
}
