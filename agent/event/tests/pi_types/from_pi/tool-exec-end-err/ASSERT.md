## Expected
- Single AgentEvent with ActionToolCall, PhaseEnd, and non-zero exit code.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"tool_call"`)
	assertContains(t, resp.Output, `"phase":"end"`)
	assertContains(t, resp.Output, `"output":"file not found"`)
}
```
