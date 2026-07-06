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