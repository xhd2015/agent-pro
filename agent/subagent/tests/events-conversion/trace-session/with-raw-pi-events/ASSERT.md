## Expected
- **After coalescing fix**: consecutive message deltas are merged into a single ASSISTANT block.
- "Hello" appears exactly once (coalesced, no duplication).
- The coalesced text contains all deltas concatenated: "Hello world from pi".
- Only 1 ASSISTANT header appears (not 3).

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    // Event count in header shows raw file line count (unchanged)
    if !strings.Contains(resp.Stdout, "3 lines") {
        t.Fatalf("expected event count 3 in output, got:\n%s", resp.Stdout)
    }

    // After fix: all deltas coalesced into 1 block, "Hello" appears once
    helloCount := strings.Count(resp.Stdout, "Hello")
    if helloCount > 1 {
        t.Fatalf("DUPLICATION: 'Hello' appears %d times (expected 1), output:\n%s", helloCount, resp.Stdout)
    }
    // Only 1 ASSISTANT header for the coalesced block
    assistantCount := strings.Count(resp.Stdout, "ASSISTANT")
    if assistantCount != 1 {
        t.Fatalf("expected exactly 1 ASSISTANT block (coalesced), got %d, output:\n%s", assistantCount, resp.Stdout)
    }
    // The coalesced text contains all words in sequence
    if !strings.Contains(resp.Stdout, "Hello world from pi") {
        t.Fatalf("expected coalesced text 'Hello world from pi' in output, got:\n%s", resp.Stdout)
    }
}
```
