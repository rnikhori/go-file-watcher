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

	target := filepath.Join(dir, "main.go")

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	// The watcher needs a moment to finish walking the directory and
	// register with the OS before it will see events, and that delay is
	// system-load dependent. Keep writing the target file until the
	// process exits (because it finally saw the event) rather than
	// gambling on a single fixed sleep.
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = os.WriteFile(target, []byte("package main"), 0o644)
			case <-stop:
				return
			}
		}
	}()

	select {
	case err := <-done:
		close(stop)
		if err != nil {
			t.Fatalf("process exited with error: %v\nstderr: %s", err, stderr.String())
		}
	case <-time.After(15 * time.Second):
		close(stop)
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
