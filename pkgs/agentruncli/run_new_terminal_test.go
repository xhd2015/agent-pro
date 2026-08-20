package agentruncli

import (
	"slices"
	"strings"
	"testing"

	flags "github.com/xhd2015/less-flags"
)

func TestReconstructRunRemainder_stripsNewTerminalKeepsBinaryAndPrompt(t *testing.T) {
	var recorded flags.Flags
	var open, nt, fromPrompt bool
	var binary, configHome string
	var env []string
	remain, err := flags.Bool("--open", &open).
		Bool("--new-terminal", &nt).
		Bool("--session-id-from-prompt", &fromPrompt).
		String("--agent-runner-binary", &binary).
		String("--agent-runner-config-home", &configHome).
		StringSlice("-e,--env", &env).
		CollectParsedFlags(&recorded).
		Parse([]string{
			"--open", "--new-terminal", "--session-id-from-prompt",
			"--agent-runner-binary", "llm-mock-run-codex",
			"--agent-runner-config-home", "/tmp/codex-home",
			"--env", "LLM_MOCK_MCP=slow_01=1s-10s",
			"--", "wait then say done",
		})
	if err != nil {
		t.Fatal(err)
	}
	got := reconstructRunRemainder(recorded, remain, "")
	joined := strings.Join(got, "\x00")
	if slices.Contains(got, "--new-terminal") {
		t.Fatalf("child argv still has --new-terminal: %#v", got)
	}
	if got[0] != "run" {
		t.Fatalf("got[0]=%q", got[0])
	}
	if !slices.Contains(got, "--open") || !slices.Contains(got, "--session-id-from-prompt") {
		t.Fatalf("missing --open / --session-id-from-prompt: %#v", got)
	}
	if !slices.Contains(got, "llm-mock-run-codex") || !slices.Contains(got, "/tmp/codex-home") {
		t.Fatalf("missing binary/config-home: %#v", got)
	}
	if !strings.Contains(joined, "LLM_MOCK_MCP=slow_01=1s-10s") {
		t.Fatalf("missing MCP env: %#v", got)
	}
	if !slices.Contains(got, "--") || !slices.Contains(got, "wait then say done") {
		t.Fatalf("missing prompt after --: %#v", got)
	}
}

func TestReconstructRunRemainder_promptFileOmitsRemain(t *testing.T) {
	var recorded flags.Flags
	var nt bool
	var pf string
	remain, err := flags.Bool("--new-terminal", &nt).
		String("--prompt-file", &pf).
		CollectParsedFlags(&recorded).
		Parse([]string{"--new-terminal", "--prompt-file", "/tmp/p.txt"})
	if err != nil {
		t.Fatal(err)
	}
	got := reconstructRunRemainder(recorded, remain, pf)
	if slices.Contains(got, "--") {
		t.Fatalf("prompt-file child should not add -- remain: %#v", got)
	}
	if !slices.Contains(got, "/tmp/p.txt") {
		t.Fatalf("missing prompt-file: %#v", got)
	}
}
