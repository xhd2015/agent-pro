## Expected
- One canonical AgentEvent with type `done`.
- Output reflects the cancelled state.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"done"`)
}
```
