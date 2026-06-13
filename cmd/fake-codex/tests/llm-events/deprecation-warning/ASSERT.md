## Expected
- The command succeeds (stdout_events still works).
- stdout contains the expected codex events.
- stderr contains a deprecation warning.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"deprecated codex works"`)
    assertContains(t, resp.Stderr, "stdout_events")
    assertContains(t, resp.Stderr, "deprecat")
}
```
