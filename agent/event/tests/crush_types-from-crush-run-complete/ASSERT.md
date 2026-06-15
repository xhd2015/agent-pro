## Expected
- One canonical AgentEvent with type `done`.
- Text contains the run output text.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"done"`)
	assertContains(t, resp.Output, `"success output"`)
}
```
