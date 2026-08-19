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

func TestSuggestFromProviderExit_alreadyRunning(t *testing.T) {
	msg := "Error: Failed to resume session from /x/rollout.jsonl: thread/resume failed during TUI bootstrap: thread/resume failed: thread 01a018ff-2e43-75a3-acd8-60847be8caa3 is already running"
	hint := suggestFromProviderExit("codex-tty", msg, "")
	if hint == "" {
		t.Fatal("expected hint")
	}
	if !strings.Contains(strings.ToLower(hint), "already running") {
		t.Fatalf("hint should say already running: %q", hint)
	}
	if !strings.Contains(strings.ToLower(hint), "another terminal") {
		t.Fatalf("hint should point at the other terminal: %q", hint)
	}
	if strings.Contains(strings.ToLower(hint), "codex resume") {
		t.Fatalf("hint must not suggest codex resume: %q", hint)
	}
	if strings.Contains(strings.ToLower(hint), "takeover") {
		t.Fatalf("hint must not suggest takeover: %q", hint)
	}
}

func TestSuggestFromProviderExit_activeWriter(t *testing.T) {
	msg := "Error: Failed to resume session: failed to acquire thread writer lock: thread 01a018ff already has an active writer"
	hint := suggestFromProviderExit("codex-tty", msg, "")
	if !strings.Contains(strings.ToLower(hint), "already running") {
		t.Fatalf("hint should say already running: %q", hint)
	}
	if strings.Contains(strings.ToLower(hint), "codex resume") || strings.Contains(strings.ToLower(hint), "takeover") {
		t.Fatalf("hint must not suggest resume/takeover: %q", hint)
	}
}

func TestSuggestFromProviderExit_threadResumeFailedTruncated(t *testing.T) {
	// TUI-chopped UUID with no "already running" suffix still matches.
	msg := "Error: Failed to resume session from /x/rollout.jsonl: thread/resume failed during TUI bootstrap: thread/resume failed: thread 01a018ff-2e43-75a"
	hint := suggestFromProviderExit("codex-tty", msg, "")
	if !strings.Contains(strings.ToLower(hint), "already running") {
		t.Fatalf("truncated thread/resume failed should still hint live thread: %q", hint)
	}
	if strings.Contains(strings.ToLower(hint), "codex resume") || strings.Contains(strings.ToLower(hint), "takeover") {
		t.Fatalf("hint must not suggest resume/takeover: %q", hint)
	}
}

func TestSuggestFromProviderExit_genericResume(t *testing.T) {
	msg := "Error: Failed to resume session from /x/rollout.jsonl: missing rollout"
	hint := suggestFromProviderExit("codex-tty", msg, "")
	if !strings.Contains(hint, "codex resume failed for this runner_session_id") {
		t.Fatalf("generic resume should keep existing hint: %q", hint)
	}
}

func TestExtractProviderExitMessage_wrappedAlreadyRunning(t *testing.T) {
	sb := "Error: Failed to resume session from /x/rollout.jsonl: thread/resume failed: thread 01a018ff-2e43-75a\nis already running\n[Terminal exited]\n"
	got := extractProviderExitMessage(sb)
	if !strings.Contains(strings.ToLower(got), "already running") {
		t.Fatalf("want wrapped 'already running' glued onto Error: line, got %q", got)
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
