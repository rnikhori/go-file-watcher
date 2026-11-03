package config

import (
	"reflect"
	"testing"
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

func TestParseInvalidFlag(t *testing.T) {
	if _, err := Parse([]string{"--not-a-real-flag"}); err == nil {
		t.Fatal("expected error for unknown flag")
	}
}
