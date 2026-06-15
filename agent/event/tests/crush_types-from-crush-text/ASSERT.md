## Expected
- One canonical AgentEvent with type `message`.
- Text contains the message content.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"message"`)
	assertContains(t, resp.Output, `"hello world"`)
}
```
