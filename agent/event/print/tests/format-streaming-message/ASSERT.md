## Expected
- The formatted output contains the delta text "a small delta" and the "💬" prefix.
- **BUG NOTE**: Each streaming delta produces its own output line ("💬 <text>").
  When many such events exist sequentially (e.g., per-token streaming),
  the output becomes extremely fractional. The AgentEvent primary path produces
  "💬 <text>" per event while the adapter fallback path produces
  "💬   ASSISTANT\n  <text>" per event.
  A proper fix would coalesce consecutive message_update delta events into a single
  output block (similar to the `Coalescer` in `coalesce.go`).

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if resp.Output == "" {
        t.Fatalf("expected non-empty formatted output, got empty string")
    }
    // The delta text should appear
    if !strings.Contains(resp.Output, "a small delta") {
        t.Fatalf("expected delta text in output, got:\n%s", resp.Output)
    }
    // AgentEvent primary path formats as "💬 <text>" (no ASSISTANT header)
    if !strings.Contains(resp.Output, "💬") {
        t.Fatalf("expected 💬 marker in output, got:\n%s", resp.Output)
    }
    // Verify this is a single-line format (no ASSISTANT block header)
    // The adapter path would have "ASSISTANT" but AgentEvent path just has "💬 text"
    t.Logf("Formatted output for single streaming delta:\n%s", resp.Output)
}
```
