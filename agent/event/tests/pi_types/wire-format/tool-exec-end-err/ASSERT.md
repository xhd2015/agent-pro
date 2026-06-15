## Expected
- Output contains isError:true and the error result.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"tool_execution_end"`)
	assertContains(t, resp.Output, `"isError":true`)
	assertContains(t, resp.Output, `"file not found"`)
}
```
