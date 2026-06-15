## Expected
- Single AgentEvent with ActionToolCall and PhaseStart.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"tool_call"`)
	assertContains(t, resp.Output, `"phase":"start"`)
}
```
