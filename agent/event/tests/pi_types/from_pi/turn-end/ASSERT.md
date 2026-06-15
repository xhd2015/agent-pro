## Expected
- Single AgentEvent with ActionStepFinish and PhaseEnd.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"step_finish"`)
	assertContains(t, resp.Output, `"phase":"end"`)
}
```
