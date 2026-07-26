## Expected
- The output lists a session with ID `explicit_test_123`.
- The session is found in the custom base directory.
- No error is returned.

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatalf("run failed: %v", err)
    }

    if !strings.Contains(resp.Stdout, "explicit_test_123") {
        t.Fatalf("expected session 'explicit_test_123' in output, got:\n%s", resp.Stdout)
    }
}
```
