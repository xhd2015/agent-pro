## Expected
- One crush event: type `agent_event` with nested type `error`.
- Contains the error text.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"agent_event"`)
	assertContains(t, resp.Output, `"type":"error"`)
	assertContains(t, resp.Output, `"something went wrong"`)
}
```
