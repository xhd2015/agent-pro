## Expected
- Single AgentEvent with ActionStepStart and PhaseStart.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"step_start"`)
	assertContains(t, resp.Output, `"phase":"start"`)
}
```
