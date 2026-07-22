## Expected

- `Stats` returns populated `SessionStats` with identity from `summary.json`.
- Counts and session latency match `signals.json` counters.
- `ToolCompleted` equals the number of `tool_completed` event lines (4).
- `Tools` includes `read_file` (3 success, avg/med 20, min 10, max 30) and
  `bash` (1 error) with correct duration aggregates.
- `ThinkingBlocks` is 2 (one coalesced run of two chunks, plus one after gap).
- `BackgroundTasks` and `Subagents` are non-nil with expected aggregates.
- All four `Sources` flags are true; warnings empty or free of missing-file noise.
- `FormatStatsText` includes session id, title, Counts, and Sources.

## Errors

- None.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if resp.Stats == nil {
		t.Fatal("stats is nil")
	}
	st := resp.Stats

	if st.ID != req.SessionID {
		t.Fatalf("ID = %q, want %q", st.ID, req.SessionID)
	}
	if st.Title != "Analyse session stats" {
		t.Fatalf("Title = %q, want Analyse session stats", st.Title)
	}
	if st.Model != "grok-composer-2.5-fast" {
		t.Fatalf("Model = %q, want grok-composer-2.5-fast", st.Model)
	}
	if st.Agent != "cursor" {
		t.Fatalf("Agent = %q, want cursor", st.Agent)
	}

	if st.Turns != 2 || st.UserMessages != 2 || st.AssistantMessages != 5 {
		t.Fatalf("turn/user/assistant = %d/%d/%d, want 2/2/5",
			st.Turns, st.UserMessages, st.AssistantMessages)
	}
	if st.ToolCalls != 3 || st.ToolFailed != 1 || st.Compactions != 1 {
		t.Fatalf("toolCalls/failed/compactions = %d/%d/%d, want 3/1/1",
			st.ToolCalls, st.ToolFailed, st.Compactions)
	}
	if st.Cancellations != 0 || st.Errors != 0 {
		t.Fatalf("cancellations/errors = %d/%d, want 0/0", st.Cancellations, st.Errors)
	}
	if st.SessionDurationSec != 120 || st.AvgResponseMs != 1500 || st.AvgTTFTMs != 400 {
		t.Fatalf("latency = %ds / %dms / %dms TTFT, want 120/1500/400",
			st.SessionDurationSec, st.AvgResponseMs, st.AvgTTFTMs)
	}

	if st.ToolCompleted != 4 {
		t.Fatalf("ToolCompleted = %d, want 4", st.ToolCompleted)
	}
	if st.ThinkingBlocks != 2 {
		t.Fatalf("ThinkingBlocks = %d, want 2", st.ThinkingBlocks)
	}

	var foundRead, foundBash bool
	for _, tool := range st.Tools {
		switch tool.Name {
		case "read_file":
			foundRead = true
			if tool.Count != 3 || tool.Success != 3 || tool.Error != 0 {
				t.Fatalf("read_file counts = N=%d S=%d E=%d, want 3/3/0",
					tool.Count, tool.Success, tool.Error)
			}
			assertFloatNear(t, "read_file.AvgMs", tool.AvgMs, 20, 0.01)
			assertFloatNear(t, "read_file.MedMs", tool.MedMs, 20, 0.01)
			assertFloatNear(t, "read_file.MinMs", tool.MinMs, 10, 0.01)
			assertFloatNear(t, "read_file.MaxMs", tool.MaxMs, 30, 0.01)
		case "bash":
			foundBash = true
			if tool.Count != 1 || tool.Success != 0 || tool.Error != 1 {
				t.Fatalf("bash counts = N=%d S=%d E=%d, want 1/0/1",
					tool.Count, tool.Success, tool.Error)
			}
			assertFloatNear(t, "bash.AvgMs", tool.AvgMs, 100, 0.01)
		}
	}
	if !foundRead {
		t.Fatalf("Tools missing read_file: %+v", st.Tools)
	}
	if !foundBash {
		t.Fatalf("Tools missing bash: %+v", st.Tools)
	}

	if st.BackgroundTasks == nil {
		t.Fatal("BackgroundTasks is nil")
	}
	if st.BackgroundTasks.Count != 1 {
		t.Fatalf("BackgroundTasks.Count = %d, want 1", st.BackgroundTasks.Count)
	}
	assertFloatNear(t, "BackgroundTasks.AvgMs", st.BackgroundTasks.AvgMs, 5000, 0.01)
	assertFloatNear(t, "BackgroundTasks.MaxMs", st.BackgroundTasks.MaxMs, 5000, 0.01)

	if st.Subagents == nil {
		t.Fatal("Subagents is nil")
	}
	if st.Subagents.Count != 1 {
		t.Fatalf("Subagents.Count = %d, want 1", st.Subagents.Count)
	}
	assertFloatNear(t, "Subagents.AvgMs", st.Subagents.AvgMs, 5000, 0.01)
	assertFloatNear(t, "Subagents.MaxMs", st.Subagents.MaxMs, 5000, 0.01)

	if !st.Sources.Summary || !st.Sources.Signals || !st.Sources.Events || !st.Sources.Updates {
		t.Fatalf("Sources flags = summary=%v signals=%v events=%v updates=%v, want all true",
			st.Sources.Summary, st.Sources.Signals, st.Sources.Events, st.Sources.Updates)
	}

	assertContains(t, resp.Output, "Session: "+req.SessionID)
	assertContains(t, resp.Output, "Title:")
	assertContains(t, resp.Output, "Analyse session stats")
	assertContains(t, resp.Output, "Counts")
	assertContains(t, resp.Output, "Sources")
}
```
