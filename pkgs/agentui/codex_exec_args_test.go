package agentui

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildCodexExecArgs_Parity(t *testing.T) {
	ws := t.TempDir()
	abs, err := filepath.Abs(ws)
	if err != nil {
		t.Fatal(err)
	}
	args := buildCodexExecArgs(ws, "do the thing", "gpt-5.6-luna", "max")
	joined := strings.Join(args, "\x00")
	for _, want := range []string{
		"exec",
		"--json",
		"--skip-git-repo-check",
		"--cd",
		abs,
		"--sandbox",
		"danger-full-access",
		"--model",
		"gpt-5.6-luna",
		"model_reasoning_effort=max",
		`projects."` + abs + `".trust_level=trusted`,
		"do the thing",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("args missing %q:\n%#v", want, args)
		}
	}
}
