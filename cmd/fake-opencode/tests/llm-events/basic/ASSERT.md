## Expected
- The command succeeds.
- stdout contains an opencode reasoning event with session ID and text.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"reasoning"`)
    assertContains(t, resp.Stdout, `"sessionID":"sess_basic"`)
    assertContains(t, resp.Stdout, `"thinking about it"`)
}
```
