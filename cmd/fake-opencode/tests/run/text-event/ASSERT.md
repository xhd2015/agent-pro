## Expected
- The text event is emitted as JSONL with session ID.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"text"`)
    assertContains(t, resp.Stdout, `"sessionID":"sess_text"`)
    assertContains(t, resp.Stdout, `"fake opencode answered"`)
}
```

