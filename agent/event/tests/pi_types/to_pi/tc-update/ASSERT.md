## Expected
- Single tool_execution_update event.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"tool_execution_update"`)
	assertNotContains(t, resp.Output, `"type":"tool_execution_start"`)
	assertNotContains(t, resp.Output, `"type":"tool_execution_end"`)
}
```
