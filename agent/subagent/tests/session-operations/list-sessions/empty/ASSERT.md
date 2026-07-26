## Expected
- Output contains "No sessions found".

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("run failed: %v", err)
    }

    if !strings.Contains(resp.Stdout, "No sessions found") {
        t.Fatalf("expected 'No sessions found' in output, got:\n%s", resp.Stdout)
    }
}
```
