## Expected
- One AgentEvent with type `done`.
- Session ID is captured in the output (e.g., in `tool_input` or `text` field).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"done"`)
	assertContains(t, resp.Output, `"grok_session_42"`)
}
```
