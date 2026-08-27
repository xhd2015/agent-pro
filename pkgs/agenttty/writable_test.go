package agenttty

import (
	"os"
	"testing"
)

func TestDetectGrokScreenStatus_modernIdleBoxedComposer(t *testing.T) {
	scrollback := []byte("" +
		"pong\n" +
		"    Worked for 6.0s\n" +
		" │ ❯                                                        │\n" +
		" Grok 4.6 (high) · always-approve\n" +
		" Shift+Tab:mode\n")
	got := detectGrokScreenStatus(scrollback)
	if got != "idle" {
		t.Fatalf("detectGrokScreenStatus=%q want idle", got)
	}
	if DetectInputBox(string(scrollback)) != InputBoxEmpty {
		t.Fatalf("DetectInputBox=%q want empty", DetectInputBox(string(scrollback)))
	}
}

func TestDetectCodexScreenStatus_liveIdleEmptyComposer(t *testing.T) {
	// Crime scene idle-probe-10s-verify-v10: finished › + medium · glue.
	scrollback := []byte("" +
		"⚠ MCP startup incomplete (failed: computer-use)\n" +
		"› reply with exactly: pong\n" +
		"› Run /review on my current changesgpt-5.6-terra medium · ~\n")
	got := detectCodexScreenStatus(scrollback)
	if got != "idle" {
		t.Fatalf("detectCodexScreenStatus=%q want idle", got)
	}
	if DetectInputBox(string(scrollback)) != InputBoxEmpty {
		t.Fatalf("DetectInputBox=%q want empty", DetectInputBox(string(scrollback)))
	}
	st := checkCodexWritable(scrollback)
	if !st.Ready || st.State != "idle" {
		t.Fatalf("checkCodexWritable ready=%v state=%q want idle", st.Ready, st.State)
	}
}

func TestDetectInputBox_grokBoxedDraftStillOccupied(t *testing.T) {
	scrollback := " │ ❯ leftover                                                │\n Shift+Tab:mode\n"
	if DetectInputBox(scrollback) != InputBoxOccupied {
		t.Fatalf("DetectInputBox=%q want occupied", DetectInputBox(scrollback))
	}
}

func TestCheckGrokWritable_realTUIHeavyPrompt(t *testing.T) {
	scrollback := []byte("     ❯ one word of France captial\n     Turn completed in 1.6s.\n  │ ❯                                                        │\n  Mock Model · always-approve")
	st := checkGrokWritable(scrollback)
	if !st.Ready {
		t.Fatalf("expected ready after post-turn ❯ prompt, got state=%q reason=%q", st.State, st.Reason)
	}
}

func TestCheckGrokWritable_legacySingleAnglePromptUnknown(t *testing.T) {
	// Legacy Grok › chrome is no longer a writable idle signal; section parse only.
	scrollback := []byte("Grok › prompt\nResponse: hi")
	st := checkGrokWritable(scrollback)
	if st.Ready || st.State != "unknown" {
		t.Fatalf("legacy › chrome want unknown/not-ready, got ready=%v state=%q reason=%q", st.Ready, st.State, st.Reason)
	}
}

func TestCheckCodexWritable_doubleAnglePromptIdle(t *testing.T) {
	scrollback := []byte("OpenAI Codex (v0.146.0)\n• You have 2 usage limit resets available\n» Explain this codebase\n")
	st := checkCodexWritable(scrollback)
	if !st.Ready || st.State != "idle" {
		t.Fatalf("expected ready idle for » prompt, got ready=%v state=%q reason=%q", st.Ready, st.State, st.Reason)
	}
	if !hasCodexPromptMarker(string(scrollback)) {
		t.Fatal("hasCodexPromptMarker should accept » (U+00BB)")
	}
}

func TestCheckCodexWritable_doubleAngleMCPIncompleteIdle(t *testing.T) {
	scrollback := []byte("MCP startup incomplete (failed: computer-use)\n» Summarize recent commits\n")
	st := checkCodexWritable(scrollback)
	if !st.Ready || st.State != "idle" {
		t.Fatalf("expected ready idle for MCP incomplete + », got ready=%v state=%q reason=%q", st.Ready, st.State, st.Reason)
	}
}

func TestCheckGrokWritable_busyWhenThinking(t *testing.T) {
	scrollback := []byte("  │ ❯                                                        │\nthinking about your request")
	st := checkGrokWritable(scrollback)
	if st.Ready {
		t.Fatal("expected not ready while thinking")
	}
	if st.State != "busy" {
		t.Fatalf("expected busy state, got %q", st.State)
	}
}

func TestCheckGrokWritable_recapExpandThinkingIdle(t *testing.T) {
	// Crime scene 01a03d6f: post-turn Recap + Ctrl+e:expand thinking footer must be idle.
	scrollback, err := os.ReadFile("testdata/grok-writable/grok-after_recap-expand-thinking-idle-01a03d6f.txt")
	if err != nil {
		t.Fatal(err)
	}
	st := checkGrokWritable(scrollback)
	if !st.Ready || st.State != "idle" {
		t.Fatalf("desired idle after Recap + expand thinking footer; got ready=%v state=%q reason=%q",
			st.Ready, st.State, st.Reason)
	}
	if got := detectGrokScreenStatus(scrollback); got != "idle" {
		t.Fatalf("detectGrokScreenStatus=%q want idle", got)
	}
	if DetectInputBox(string(scrollback)) != InputBoxEmpty {
		t.Fatalf("DetectInputBox=%q want empty (Build anything is placeholder)", DetectInputBox(string(scrollback)))
	}
}

func TestCheckCodexWritable_notReadyAfterExitFooter(t *testing.T) {
	// Residual composer glyphs + real /exit footer must not stay sendable.
	scrollback := []byte("" +
		"› old turn\n" +
		"Token usage: total=1\n" +
		"To continue this session, run codex resume 019fdca1-3893-7fa3-a8aa-ebc1ccc750a0\n")
	st := checkCodexWritable(scrollback)
	if st.Ready {
		t.Fatalf("expected not ready after exit footer, got ready state=%q reason=%q", st.State, st.Reason)
	}
	if st.State != "exited" {
		t.Fatalf("expected state=exited, got %q reason=%q", st.State, st.Reason)
	}
}

func TestCheckCodexWritable_notReadyAfterTerminalExited(t *testing.T) {
	scrollback := []byte("› leftover\n[Terminal exited]\n$ ")
	st := checkCodexWritable(scrollback)
	if st.Ready {
		t.Fatalf("expected not ready after [Terminal exited], got ready state=%q", st.State)
	}
}

func TestCheckCodexWritable_standaloneResumeCmdStillIdle(t *testing.T) {
	// User typed "codex resume …" in chat without exit phrase — still injectable.
	scrollback := []byte("» how do I use codex resume abcdef01-2345-6789-abcd-ef0123456789\n")
	st := checkCodexWritable(scrollback)
	if !st.Ready {
		t.Fatalf("resume cmd alone should not block writable, got state=%q reason=%q", st.State, st.Reason)
	}
}