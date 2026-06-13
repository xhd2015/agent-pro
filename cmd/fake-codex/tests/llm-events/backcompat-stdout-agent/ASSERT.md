## Expected
- The command succeeds.
- stdout contains codex reasoning events with the expected text.
- stderr contains deprecation warning.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"item.started"`)
    assertContains(t, resp.Stdout, `"type":"item.completed"`)
    assertContains(t, resp.Stdout, `"agent format backcompat"`)
    assertContains(t, resp.Stderr, "deprecat")
}
```
