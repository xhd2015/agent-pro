---
label: e2e
---

## Expected
- The text event is emitted with the original text `"plain text message"`.
- The error event is emitted with the original error `"an error occurred"`.
- No tool execution occurs (no bash/read/write/grep side effects).

```go
import (
    "strings"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    stdout := resp.Stdout
    if !strings.Contains(stdout, "plain text message") {
        t.Fatalf("expected text 'plain text message' in output, got:\n%s", stdout)
    }
    if !strings.Contains(stdout, "an error occurred") {
        t.Fatalf("expected error message 'an error occurred' in output, got:\n%s", stdout)
    }
    // Verify it's valid JSONL
    events := parseJSONLines(t, stdout)
    if len(events) < 2 {
        t.Fatalf("expected at least 2 events, got %d", len(events))
    }
}
```
