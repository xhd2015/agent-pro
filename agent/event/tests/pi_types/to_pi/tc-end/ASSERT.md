## Expected
- Single tool_execution_end event with toolName and result.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"tool_execution_end"`)
	assertContains(t, resp.Output, `"toolName":"bash"`)
	assertNotContains(t, resp.Output, `"type":"tool_execution_start"`)
}
```
