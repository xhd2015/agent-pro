## Expected

- `ToolCompleted` is 4 (all tool_completed lines).
- Without signals, `ToolCalls` falls back toward tool_started count (4).
- `read_file`: Count=3, Success=3, Error=0, Avg=20, Med=20, Min=10, Max=30.
- `bash`: Count=1, Success=0, Error=1, Avg=Med=Min=Max=50.
- `Sources.Events` true; `Sources.Signals` false; warnings mention signals.

## Errors

- None.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if resp.Stats == nil {
		t.Fatal("stats is nil")
	}
	st := resp.Stats

	if st.ToolCompleted != 4 {
		t.Fatalf("ToolCompleted = %d, want 4", st.ToolCompleted)
	}
	if st.ToolCalls != 4 {
		t.Fatalf("ToolCalls = %d, want 4 (tool_started fallback without signals)", st.ToolCalls)
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
			assertFloatNear(t, "bash.AvgMs", tool.AvgMs, 50, 0.01)
			assertFloatNear(t, "bash.MedMs", tool.MedMs, 50, 0.01)
			assertFloatNear(t, "bash.MinMs", tool.MinMs, 50, 0.01)
			assertFloatNear(t, "bash.MaxMs", tool.MaxMs, 50, 0.01)
		}
	}
	if !foundRead {
		t.Fatalf("Tools missing read_file: %+v", st.Tools)
	}
	if !foundBash {
		t.Fatalf("Tools missing bash: %+v", st.Tools)
	}

	if !st.Sources.Summary || !st.Sources.Events {
		t.Fatalf("Sources summary/events = %v/%v, want true/true",
			st.Sources.Summary, st.Sources.Events)
	}
	if st.Sources.Signals {
		t.Fatal("Sources.Signals = true, want false")
	}
	warn := strings.ToLower(joinWarnings(st.Sources.Warnings))
	if !strings.Contains(warn, "signals") {
		t.Fatalf("warnings %q should mention missing signals", st.Sources.Warnings)
	}
}
```
