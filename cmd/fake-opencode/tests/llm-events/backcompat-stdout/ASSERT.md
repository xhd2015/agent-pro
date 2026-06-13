## Expected
- The command succeeds.
- stdout contains the converted event.
- stderr contains deprecation warning (stdout_events is deprecated).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"reasoning"`)
    assertContains(t, resp.Stdout, `"backward compat works"`)
    assertContains(t, resp.Stderr, "deprecat")
}
```
