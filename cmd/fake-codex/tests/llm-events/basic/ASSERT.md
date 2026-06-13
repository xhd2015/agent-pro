## Expected
- The command succeeds.
- stdout contains codex reasoning events (item.started and item.completed).
- The reasoning text appears in the completed event.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"item.started"`)
    assertContains(t, resp.Stdout, `"type":"reasoning"`)
    assertContains(t, resp.Stdout, `"type":"item.completed"`)
    assertContains(t, resp.Stdout, `"thinking in codex"`)
}
```
