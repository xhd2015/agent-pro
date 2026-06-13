## Expected
- The command succeeds.
- stdout contains a codex message completed event with the configured text.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"item.completed"`)
    assertContains(t, resp.Stdout, `"type":"message"`)
    assertContains(t, resp.Stdout, `"task completed"`)
}
```
