## Expected
- One canonical AgentEvent with type `done`.
- Text contains the error message from run_complete.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"done"`)
	assertContains(t, resp.Output, `"agent run failed"`)
}
```
