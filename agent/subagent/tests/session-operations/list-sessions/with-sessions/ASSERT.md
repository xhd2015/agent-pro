## Expected
- Output lists both `session_alpha` and `session_beta`.
- `session_alpha` (newer) appears before `session_beta` (older).

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("run failed: %v", err)
    }

    if !strings.Contains(resp.Stdout, "session_alpha") {
        t.Fatalf("expected 'session_alpha' in output, got:\n%s", resp.Stdout)
    }
    if !strings.Contains(resp.Stdout, "session_beta") {
        t.Fatalf("expected 'session_beta' in output, got:\n%s", resp.Stdout)
    }

    idxA := strings.Index(resp.Stdout, "session_alpha")
    idxB := strings.Index(resp.Stdout, "session_beta")
    if idxA == -1 || idxB == -1 {
        return
    }
    if idxA >= idxB {
        t.Fatalf("expected 'session_alpha' before 'session_beta' (sorted by time desc), got:\n%s", resp.Stdout)
    }
}
```
