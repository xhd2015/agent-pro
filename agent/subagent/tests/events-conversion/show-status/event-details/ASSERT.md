## Expected
- Output shows event count and contains text from the last event.
- Status output contains session and event information.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !strings.Contains(resp.Stdout, "4 lines") {
        t.Fatalf("expected event count 4 in output, got:\n%s", resp.Stdout)
    }
    if !strings.Contains(resp.Stdout, "event4 final") {
        t.Fatalf("expected last event text in output, got:\n%s", resp.Stdout)
    }
}
```
