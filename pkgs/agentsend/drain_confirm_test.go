package agentsend

import "testing"

func TestCodexSubmitConfirmOpts_NoSubmitSkips(t *testing.T) {
	opts := codexSubmitConfirmOpts(Session{Runner: "codex-tty", Home: t.TempDir()}, "hello", false)
	if opts != nil {
		t.Fatalf("no-submit must not confirm, got %+v", opts)
	}
}

func TestCodexSubmitConfirmOpts_GrokSkips(t *testing.T) {
	opts := codexSubmitConfirmOpts(Session{Runner: "grok-tty", Home: t.TempDir()}, "hello", true)
	if opts != nil {
		t.Fatalf("grok must not confirm, got %+v", opts)
	}
}

func TestCodexSubmitConfirmOpts_CodexSubmit(t *testing.T) {
	opts := codexSubmitConfirmOpts(Session{
		Runner:            "codex-tty",
		Home:              t.TempDir(),
		ListenAddr:        "127.0.0.1:9",
		TerminalSessionID: "term-1",
	}, "hello from seatalk", true)
	if opts == nil {
		t.Fatal("codex submit must confirm")
	}
	if opts.Prompt != "hello from seatalk" || opts.ListenAddr != "127.0.0.1:9" || opts.TerminalSessionID != "term-1" {
		t.Fatalf("opts = %+v", opts)
	}
	if opts.Baselines == nil {
		t.Fatal("baselines must be captured, not nil")
	}
}
