## Expected
- Output header contains the session ID.
- "(no events yet)" message appears in the output.
- "Done" or "Session finished" message appears.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("run failed: %v", err)
    }

    if !strings.Contains(resp.Stdout, "trace_noevents_1") {
        t.Fatalf("expected 'trace_noevents_1' in stdout, got:\n%s", resp.Stdout)
    }

    if !strings.Contains(resp.Stdout, "no events") {
        t.Fatalf("expected 'no events' message in stdout, got:\n%s", resp.Stdout)
    }
}
```
