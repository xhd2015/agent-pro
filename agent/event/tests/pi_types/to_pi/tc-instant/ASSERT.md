## Expected
- Two pi events: tool_execution_start and tool_execution_end.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"tool_execution_start"`)
	assertContains(t, resp.Output, `"type":"tool_execution_end"`)
	assertContains(t, resp.Output, `"toolName":"bash"`)
}
```
