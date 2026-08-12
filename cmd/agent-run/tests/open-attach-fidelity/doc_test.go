package openattachfidelity_test

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

// TestOpenAttachFidelityDoctest runs the open-attach-fidelity suite.
func TestOpenAttachFidelityDoctest(t *testing.T) {
	if _, err := exec.LookPath("doctest"); err != nil {
		t.Skip("doctest not on PATH")
	}

	root := repoRoot(t)
	suite := filepath.Join(root, "cmd/agent-run/tests/open-attach-fidelity")
	cmd := exec.Command("doctest", "test", "--rm", "-count=1", "--label", "e2e", suite)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("doctest open-attach-fidelity: %v", err)
	}
}
