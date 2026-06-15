## Expected
- One crush event: payload type `message` with role `assistant`.
- Message parts contain a `reasoning` part with the think text.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"message"`)
	assertContains(t, resp.Output, `"role":"assistant"`)
	assertContains(t, resp.Output, `"type":"reasoning"`)
	assertContains(t, resp.Output, `"thinking about the problem"`)
}
```
