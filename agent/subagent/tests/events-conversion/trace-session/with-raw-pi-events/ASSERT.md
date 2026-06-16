## Expected
- **After fix: no text duplication.** The trace output shows each delta only once.
- "Hello" appears at most once (from the message_start event), not duplicated across multiple events.
- Each message_update event outputs only its delta text, not the full accumulated text.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    // Should contain event count
    if !strings.Contains(resp.Stdout, "3 lines") {
        t.Fatalf("expected event count 3 in output, got:\n%s", resp.Stdout)
    }

    // After fix: "Hello" appears exactly once (from message_start event)
    helloCount := strings.Count(resp.Stdout, "Hello")
    if helloCount > 1 {
        t.Fatalf("DUPLICATION STILL PRESENT: 'Hello' appears %d times (expected 1), output:\n%s", helloCount, resp.Stdout)
    }
    // Verify deltas appear (not full accumulated text)
    if !strings.Contains(resp.Stdout, "world") {
        t.Fatalf("expected delta 'world' in output, got:\n%s", resp.Stdout)
    }
    if !strings.Contains(resp.Stdout, "from pi") {
        t.Fatalf("expected delta 'from pi' in output, got:\n%s", resp.Stdout)
    }
    // The accumulated full text "Hello world from pi" should NOT appear
    // (only the individual deltas should be shown)
}
```
