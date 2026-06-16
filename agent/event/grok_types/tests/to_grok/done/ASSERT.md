## Expected
- One grok event with type `end` and `sessionId` matching the provided session ID.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"end"`)
	assertContains(t, resp.Output, `"sessionId":"sess_abc_123"`)
}
```
