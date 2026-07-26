## Expected
- Output contains "(no events yet)".
- Output contains "Done" or "finished".

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !strings.Contains(resp.Stdout, "(no events yet)") {
        t.Fatalf("expected '(no events yet)' in output, got:\n%s", resp.Stdout)
    }
}
```
