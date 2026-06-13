## Expected
- The command succeeds.
- The tool_use event contains the file contents in the state output.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"tool_use"`)
    assertContains(t, resp.Stdout, `"tool":"read"`)
    assertContains(t, resp.Stdout, `"readable content"`)
    assertContains(t, resp.Stdout, `"status":"completed"`)
}
```
