package weblayout_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../../../.."))
}

// TestWebLayoutDoctest runs the playwright mobile web-layout doctest suite.
func TestWebLayoutDoctest(t *testing.T) {
	if _, err := exec.LookPath("doctest"); err != nil {
		t.Skip("doctest not on PATH")
	}
	if _, err := exec.LookPath("playwright-debug"); err != nil {
		t.Skip("playwright-debug not on PATH")
	}

	root := repoRoot(t)
	suite := filepath.Join(root, "cmd/agent-run/tests/web-layout")
	cmd := exec.Command("doctest", "test", "--rm", "-count=1", suite, "--label", "chromium")
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("doctest web-layout: %v", err)
	}
}