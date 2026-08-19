package agenttty

import "testing"

// Formal desired-behavior guard for the scorer WaitDone idle false-negative.
// Historical "• Working" / "esc to interrupt" above a settled bottom › must be idle.
// RED on current checkCodexWritable (full-scrollback busy rule). Product fix later.
func TestCheckCodexWritable_HistoricalWorkingWithBottomPrompt_Idle(t *testing.T) {
	scrollback := []byte(`
• Ran tool write score.json
• Working
  esc to interrupt

some answer text here...

- Worked for 1m 42s ------------------------------------------------------------
› 
  gpt-5.6-luna max · ~/project…
`)
	st := checkCodexWritable(scrollback)
	t.Logf("Ready=%v State=%s Reason=%q", st.Ready, st.State, st.Reason)
	if !st.Ready || st.State != "idle" {
		t.Fatalf("desired: ready idle after historical Working above settled › prompt; got ready=%v state=%q reason=%q",
			st.Ready, st.State, st.Reason)
	}
}

func TestCheckCodexWritable_IdlePromptOnly(t *testing.T) {
	scrollback := []byte(`
- Worked for 1m 42s ------------------------------------------------------------
› Implement {feature}
  gpt-5.6-luna max · ~/project…
`)
	st := checkCodexWritable(scrollback)
	t.Logf("Ready=%v State=%s Reason=%q", st.Ready, st.State, st.Reason)
	if !st.Ready {
		t.Fatalf("expected idle prompt ready, got %v %q", st.Ready, st.Reason)
	}
}

// Live Working without a settled bottom prompt must remain non-ready (control: still busy).
func TestCheckCodexWritable_LiveWorkingNoSettledPrompt_Busy(t *testing.T) {
	scrollback := []byte(`
• Working
  esc to interrupt
`)
	st := checkCodexWritable(scrollback)
	t.Logf("Ready=%v State=%s Reason=%q", st.Ready, st.State, st.Reason)
	if st.Ready {
		t.Fatalf("live Working without settled › must not be ready; got ready state=%q reason=%q", st.State, st.Reason)
	}
	if st.State != "busy" {
		t.Fatalf("expected busy for live Working, got state=%q reason=%q", st.State, st.Reason)
	}
}

// Live Working immediately above a placeholder composer › must stay busy.
// Experiment snap-00: last › is "Write tests for @filename"; snapping to that
// line dropped "• Working (6s • esc to interrupt)".
func TestCheckCodexWritable_WorkingAbovePlaceholderPrompt_Busy(t *testing.T) {
	scrollback := []byte("" +
		"› run sleep 20 then say done\n" +
		"• Working (6s • esc to interrupt) · 1 background terminal running · /ps to view…\n" +
		"› Write tests for @filenamemock-model default · ~/project…\n")
	st := checkCodexWritable(scrollback)
	t.Logf("Ready=%v State=%s Reason=%q", st.Ready, st.State, st.Reason)
	if st.Ready {
		t.Fatalf("Working above placeholder › must not be ready; got ready state=%q reason=%q", st.State, st.Reason)
	}
	if st.State != "busy" {
		t.Fatalf("expected busy, got state=%q reason=%q", st.State, st.Reason)
	}
	if detectCodexScreenStatus(scrollback) != "busy" {
		t.Fatalf("DetectScreenStatus=%q want busy", detectCodexScreenStatus(scrollback))
	}
}
