## Expected
- The error message indicates session ID cannot be detected.
- The error message includes the string `gen_` (generated session ID prefix).
- The error message includes guidance to use `--session-id`.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if resp.Err != nil {
        // the Run() function may return the error via Response.Err
        if !strings.Contains(resp.Err.Error(), "cannot detect session id") &&
           !strings.Contains(resp.Err.Error(), "session") {
            t.Fatalf("expected session detection error, got: %v", resp.Err)
        }
        return
    }

    combined := resp.Stderr
    if combined == "" {
        t.Fatal("expected an error about missing session ID, but stderr was empty")
    }

    if !strings.Contains(combined, "cannot detect") && !strings.Contains(combined, "session") {
        t.Fatalf("expected session detection error in stderr, got:\n%s", combined)
    }
}
```
