## Expected
- The command succeeds.
- stdout contains the event output.
- stderr does NOT contain any deprecation warning.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"clean"`)
    if strings.Contains(resp.Stderr, "deprecat") || strings.Contains(resp.Stderr, "stdout_events") {
        t.Fatalf("unexpected deprecation warning in stderr:\n%s", resp.Stderr)
    }
}
```
