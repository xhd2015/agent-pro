package agenttty

import "testing"

func TestCheckGrokWritable_realTUIHeavyPrompt(t *testing.T) {
	scrollback := []byte("     ❯ one word of France captial\n     Turn completed in 1.6s.\n  │ ❯                                                        │\n  Mock Model · always-approve")
	st := checkGrokWritable(scrollback)
	if !st.Ready {
		t.Fatalf("expected ready after post-turn ❯ prompt, got state=%q reason=%q", st.State, st.Reason)
	}
}

func TestCheckGrokWritable_legacySingleAnglePrompt(t *testing.T) {
	scrollback := []byte("Grok › prompt\nResponse: hi")
	st := checkGrokWritable(scrollback)
	if !st.Ready {
		t.Fatalf("expected ready for legacy › prompt, got state=%q reason=%q", st.State, st.Reason)
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