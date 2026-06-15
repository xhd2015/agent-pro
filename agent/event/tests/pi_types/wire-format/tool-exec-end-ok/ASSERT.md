## Expected
- Output contains tool_execution_end type, result, and isError:false.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"tool_execution_end"`)
	assertContains(t, resp.Output, `"toolCallId":"call_1"`)
	assertContains(t, resp.Output, `"isError":false`)
	assertContains(t, resp.Output, `file.txt`)
}
```
