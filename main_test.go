package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// buildBinary compiles the CLI once per test run and returns the path to
// the resulting executable.
func buildBinary(t *testing.T) string {
	t.Helper()

	bin := filepath.Join(t.TempDir(), "file-watcher")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}

	cmd := exec.Command("go", "build", "-o", bin, ".")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("building binary: %v\n%s", err, out)
	}
	return bin
}

func TestCLIOnceJSONEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping end-to-end build/run test in -short mode")
	}

	bin := buildBinary(t)
	dir := t.TempDir()

	cmd := exec.Command(bin, "--once", "--json", "--ext=.go", dir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("starting command: %v", err)
	}

	time.Sleep(1 * time.Second)
	target := filepath.Join(dir, "main.go")
	if err := os.WriteFile(target, []byte("package main"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("process exited with error: %v\nstderr: %s", err, stderr.String())
		}
	case <-time.After(10 * time.Second):
		cmd.Process.Kill()
		t.Fatalf("process did not exit after --once matched; stderr: %s", stderr.String())
	}

	line := strings.TrimSpace(stdout.String())
	if line == "" {
		t.Fatalf("expected one JSON line on stdout, got none; stderr: %s", stderr.String())
	}

	var ev struct {
		Time string `json:"time"`
		Op   string `json:"op"`
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(line), &ev); err != nil {
		t.Fatalf("stdout line is not valid JSON: %v\nline: %q", err, line)
	}
	if ev.Path != target {
		t.Errorf("Path = %q, want %q", ev.Path, target)
	}
}
