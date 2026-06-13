## Expected
- The command succeeds.
- stdout contains reasoning, text, and done event types with the expected text.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"reasoning"`)
    assertContains(t, resp.Stdout, `"type":"text"`)
    assertContains(t, resp.Stdout, `"type":"done"`)
    assertContains(t, resp.Stdout, `"thinking"`)
    assertContains(t, resp.Stdout, `"hello"`)
}
```
