## Expected

- `Stats` succeeds with identity and signal-derived counts/latency.
- `Sources.Summary` and `Sources.Signals` are true.
- `Sources.Events` and `Sources.Updates` are false.
- `Sources.Warnings` mentions missing `events` and `updates` (human-readable).
- Per-tool list is empty (no events); background/subagent aggregates are nil.
- ThinkingBlocks is 0 without updates.

## Errors

- None.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if resp.Stats == nil {
		t.Fatal("stats is nil")
	}
	st := resp.Stats

	if st.ID != req.SessionID {
		t.Fatalf("ID = %q, want %q", st.ID, req.SessionID)
	}
	if st.Title != "Signals only session" {
		t.Fatalf("Title = %q, want Signals only session", st.Title)
	}
	if st.Turns != 2 || st.UserMessages != 2 || st.AssistantMessages != 5 {
		t.Fatalf("turn/user/assistant = %d/%d/%d, want 2/2/5",
			st.Turns, st.UserMessages, st.AssistantMessages)
	}
	if st.ToolCalls != 3 || st.ToolFailed != 1 {
		t.Fatalf("toolCalls/failed = %d/%d, want 3/1", st.ToolCalls, st.ToolFailed)
	}
	if st.SessionDurationSec != 120 || st.AvgResponseMs != 1500 || st.AvgTTFTMs != 400 {
		t.Fatalf("latency = %d / %d / %d, want 120/1500/400",
			st.SessionDurationSec, st.AvgResponseMs, st.AvgTTFTMs)
	}

	if st.ToolCompleted != 0 {
		t.Fatalf("ToolCompleted = %d, want 0 without events", st.ToolCompleted)
	}
	if len(st.Tools) != 0 {
		t.Fatalf("Tools = %+v, want empty without events", st.Tools)
	}
	if st.ThinkingBlocks != 0 {
		t.Fatalf("ThinkingBlocks = %d, want 0 without updates", st.ThinkingBlocks)
	}
	if st.BackgroundTasks != nil {
		t.Fatalf("BackgroundTasks = %+v, want nil", st.BackgroundTasks)
	}
	if st.Subagents != nil {
		t.Fatalf("Subagents = %+v, want nil", st.Subagents)
	}

	if !st.Sources.Summary || !st.Sources.Signals {
		t.Fatalf("Sources summary/signals = %v/%v, want true/true",
			st.Sources.Summary, st.Sources.Signals)
	}
	if st.Sources.Events || st.Sources.Updates {
		t.Fatalf("Sources events/updates = %v/%v, want false/false",
			st.Sources.Events, st.Sources.Updates)
	}

	warn := strings.ToLower(joinWarnings(st.Sources.Warnings))
	if !strings.Contains(warn, "events") {
		t.Fatalf("warnings %q should mention missing events", st.Sources.Warnings)
	}
	if !strings.Contains(warn, "updates") {
		t.Fatalf("warnings %q should mention missing updates", st.Sources.Warnings)
	}
}
```
