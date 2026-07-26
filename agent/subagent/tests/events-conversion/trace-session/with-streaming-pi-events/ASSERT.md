## Expected
- **FIX**: Consecutive streaming deltas are coalesced into a single ASSISTANT block.
- There are 5 message_update deltas + 1 message_start = 6 text-producing events.
- After the fix, all 6 events produce exactly 1 ASSISTANT header (coalesced).
- The complete text "Let me start by understanding" appears in that single block.
- No fractional per-delta output.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if resp.Stdout == "" {
        t.Fatalf("expected non-empty output")
    }

    // After fix: all deltas coalesced into exactly 1 ASSISTANT block
    assistantCount := strings.Count(resp.Stdout, "ASSISTANT")
    if assistantCount != 1 {
        t.Fatalf("expected exactly 1 ASSISTANT block (coalesced), got %d, output:\n%s", assistantCount, resp.Stdout)
    }
    // The complete text from all deltas should appear
    if !strings.Contains(resp.Stdout, "Let me start by understanding") {
        t.Fatalf("expected coalesced text 'Let me start by understanding' in output, got:\n%s", resp.Stdout)
    }
    // Individual words still present (they make up the coalesced text)
    if !strings.Contains(resp.Stdout, "understanding") {
        t.Fatalf("expected delta text 'understanding' in output, got:\n%s", resp.Stdout)
    }
}
```
