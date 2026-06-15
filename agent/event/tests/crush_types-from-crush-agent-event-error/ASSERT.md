## Expected
- One canonical AgentEvent with type `error`.
- Text contains the error message.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"error"`)
	assertContains(t, resp.Output, `"runtime failure"`)
}
```
