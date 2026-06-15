## Expected
- One canonical AgentEvent with type `think`.
- Text contains the reasoning content.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"think"`)
	assertContains(t, resp.Output, `"step by step reasoning"`)
}
```
