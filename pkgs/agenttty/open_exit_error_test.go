package agenttty

import (
	"strings"
	"testing"
)

func TestExtractProviderExitMessage_errorLine(t *testing.T) {
	sb := "noise\nError: Failed to resume session from /x/rollout.jsonl: database disk image is malformed\n[Terminal exited]\n"
	got := extractProviderExitMessage(sb)
	if !strings.Contains(strings.ToLower(got), "error:") {
		t.Fatalf("want Error: line, got %q", got)
	}
	if !strings.Contains(strings.ToLower(got), "malformed") {
		t.Fatalf("want malformed, got %q", got)
	}
}

func TestSuggestFromProviderExit_malformedDB(t *testing.T) {
	msg := "Error: Failed to resume session: database disk image is malformed (thread-store)"
	hint := suggestFromProviderExit("codex-tty", msg, "")
	if hint == "" {
		t.Fatal("expected hint")
	}
	if !strings.Contains(hint, "state_5.sqlite") {
		t.Fatalf("hint should mention state_5.sqlite: %q", hint)
	}
	if !strings.Contains(hint, "Backup") && !strings.Contains(strings.ToLower(hint), "backup") && !strings.Contains(hint, "mv ") {
		t.Fatalf("hint should tell operator how to reset: %q", hint)
	}
}

func TestFormatOpenAgentExitedError_includesHint(t *testing.T) {
	// No live listen: still returns structured shell error.
	err := formatOpenAgentExitedError("codex-tty", "", "sess-x", "open attach", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	s := err.Error()
	if !strings.Contains(s, "agent exited before open attach") {
		t.Fatalf("got %q", s)
	}
}
