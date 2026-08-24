package knowledgesink

import "testing"

func TestEffectiveAgentRunner_HeadlessMapsTTY(t *testing.T) {
	got, from, err := EffectiveAgentRunner(ModeHeadless, "codex-tty")
	if err != nil || got != "codex" || from != "codex-tty" {
		t.Fatalf("got=%q from=%q err=%v", got, from, err)
	}
	got, from, err = EffectiveAgentRunner(ModeHeadless, "grok-tty")
	if err != nil || got != "grok" || from != "grok-tty" {
		t.Fatalf("got=%q from=%q err=%v", got, from, err)
	}
	got, from, err = EffectiveAgentRunner(ModeHeadless, "codex")
	if err != nil || got != "codex" || from != "" {
		t.Fatalf("got=%q from=%q err=%v", got, from, err)
	}
}

func TestEffectiveAgentRunner_OpenMapsCLI(t *testing.T) {
	got, from, err := EffectiveAgentRunner(ModeOpen, "codex")
	if err != nil || got != "codex-tty" || from != "codex" {
		t.Fatalf("got=%q from=%q err=%v", got, from, err)
	}
	got, from, err = EffectiveAgentRunner(ModeOpen, "codex-tty")
	if err != nil || got != "codex-tty" || from != "" {
		t.Fatalf("got=%q from=%q err=%v", got, from, err)
	}
}

func TestEffectiveAgentRunner_OpenRejectsNonTTY(t *testing.T) {
	_, _, err := EffectiveAgentRunner(ModeOpen, "fake-codex")
	if err == nil {
		t.Fatal("expected error")
	}
}
