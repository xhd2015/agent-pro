package grok

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/agent-pro/agent/exec"
)

func TestFindExecutableGrok_prefersFirstCandidate(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "grok-a")
	second := filepath.Join(dir, "grok-b")
	for _, p := range []string{first, second} {
		if err := os.WriteFile(p, []byte{0}, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	got, ok := findExecutableGrok([]string{first, second})
	if !ok || got != first {
		t.Fatalf("findExecutableGrok = (%q, %v), want (%q, true)", got, ok, first)
	}
}

func TestProbeGrokInstallPath_findsUnderHome(t *testing.T) {
	home := t.TempDir()
	binDir := filepath.Join(home, ".grok", "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	grokPath := filepath.Join(binDir, "grok")
	if err := os.WriteFile(grokPath, []byte{0}, 0o755); err != nil {
		t.Fatal(err)
	}
	got, ok := findExecutableGrok(grokInstallCandidates(home))
	if !ok || got != grokPath {
		t.Fatalf("findExecutableGrok = (%q, %v), want (%q, true)", got, ok, grokPath)
	}
}

func TestFindAgentPath_probesWhenPATHEmpty(t *testing.T) {
	home := t.TempDir()
	binDir := filepath.Join(home, ".grok", "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	grokPath := filepath.Join(binDir, "grok")
	if err := os.WriteFile(grokPath, []byte{0}, 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", home)
	t.Setenv("PATH", "")

	env := exec.NewEnv(&exec.PathsConfig{
		RootDirName: ".agent-pro",
		DataDirName: "data",
		BinDirName:  "bin",
	}, "AGENT_PRO_CONFIG_HOME")

	got, err := FindAgentPath(env)
	if err != nil {
		t.Fatalf("FindAgentPath: %v", err)
	}
	if got != grokPath {
		t.Fatalf("FindAgentPath = %q, want %q", got, grokPath)
	}
}