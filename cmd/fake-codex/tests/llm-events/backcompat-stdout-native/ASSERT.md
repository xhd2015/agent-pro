## Expected
- The command succeeds.
- stdout contains the native codex event (item.started directly, no conversion needed).
- stderr contains deprecation warning.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"item.started"`)
    assertContains(t, resp.Stdout, `"native format backcompat"`)
    assertContains(t, resp.Stderr, "deprecat")
}
```
