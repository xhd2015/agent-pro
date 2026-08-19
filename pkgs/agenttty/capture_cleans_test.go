package agenttty

import (
	"strings"
	"testing"
)

func TestCleanCodexScrollbackFallback_lsOutput(t *testing.T) {
	scroll := []byte("CODEX_TTY_BANNER\n>4;0m>7u\n╭────────────────────────╮\n│ >_ OpenAI Codex        │\n│ model: loading /model to change │\n│ directory: /tmp/work   │\n╰────────────────────────╯\nStarting MCP servers...\nBooting MCP...\nWorking...\nWorking...\n› run ls\nls output:\nAGENTS.md\ncmd\npkgs\n")
	got := cleanCodexScrollbackFallback(scroll, "run ls", []string{"CODEX_TTY_BANNER"})
	if !strings.Contains(got, "ls output:") || !strings.Contains(got, "AGENTS.md") {
		t.Fatalf("missing fragments in %q", got)
	}
}

func TestCleanCodexScrollbackFallback_bareCRSeparators(t *testing.T) {
	// ptywrap snapshots may separate rows with bare \r.
	scroll := []byte("CODEX_TTY_BANNER\r› run ls\rls output:\rAGENTS.md\rcmd\rpkgs\r")
	got := cleanCodexScrollbackFallback(scroll, "run ls", []string{"CODEX_TTY_BANNER"})
	if !strings.Contains(got, "ls output:") || !strings.Contains(got, "AGENTS.md") || !strings.Contains(got, "\n") {
		t.Fatalf("bare CR scrollback not split; got %q", got)
	}
}

func TestCleanCodexScrollbackFallback_gluedPromptAndResults(t *testing.T) {
	// Observed CI/local snapshot shape under KeepAlive:
	// "› run lsls output:\nAGENTS.mdcmdpkgs\n"
	scroll := []byte("CODEX_TTY_BANNER\n› run lsls output:\nAGENTS.mdcmdpkgs\n[Terminal exited]\n")
	got := cleanCodexScrollbackFallback(scroll, "run ls", []string{"CODEX_TTY_BANNER"})
	for _, want := range []string{"ls output:", "AGENTS.md", "cmd", "pkgs"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
	if strings.Contains(got, "›") || strings.Contains(got, "Terminal exited") {
		t.Fatalf("chrome leaked into %q", got)
	}
}
