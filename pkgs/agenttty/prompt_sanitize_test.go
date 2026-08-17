package agenttty

import "testing"

func TestSanitizePromptForRunner_CRAllRunners(t *testing.T) {
	t.Parallel()
	for _, runner := range []string{"codex-tty", "grok-tty", "future-tty", ""} {
		if got := SanitizePromptForRunner(runner, "a\r\nb"); got != "a\nb" {
			t.Fatalf("runner %q CRLF: got %q", runner, got)
		}
		if got := SanitizePromptForRunner(runner, "a\rb"); got != "a\nb" {
			t.Fatalf("runner %q CR: got %q", runner, got)
		}
	}
}

func TestSanitizePromptForRunner_TabCodexOnly(t *testing.T) {
	t.Parallel()
	in := "SELECT\n\tDatabase: app_orders_staging_db"
	wantCodex := "SELECT\n    Database: app_orders_staging_db"
	for _, runner := range []string{"codex-tty", "codex", "  CODEX-TTY  "} {
		got := SanitizePromptForRunner(runner, in)
		if got != wantCodex {
			t.Fatalf("runner %q: got %q want %q", runner, got, wantCodex)
		}
	}
	for _, runner := range []string{"grok-tty", "opencode-tty", "future-tty", ""} {
		got := SanitizePromptForRunner(runner, in)
		if got != in {
			t.Fatalf("runner %q should keep TAB: got %q", runner, got)
		}
	}
}

func TestSanitizePromptForRunner_Combo(t *testing.T) {
	t.Parallel()
	if got := SanitizePromptForRunner("codex-tty", "A\r\n\tB"); got != "A\n    B" {
		t.Fatalf("codex combo: %q", got)
	}
	if got := SanitizePromptForRunner("grok-tty", "A\r\n\tB"); got != "A\n\tB" {
		t.Fatalf("grok combo: %q", got)
	}
}

func TestPrepareRunnerPrompt_CodexTabThenUTF8(t *testing.T) {
	t.Parallel()
	got := prepareRunnerPrompt("codex-tty", "x\tY\rZ")
	if got != "x    Y\nZ" {
		t.Fatalf("prepare: %q", got)
	}
}
