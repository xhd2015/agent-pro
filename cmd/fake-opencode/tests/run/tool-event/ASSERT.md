## Expected
- The tool event has the correct shape: type `tool_use`, tool `bash`, status `completed`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"tool_use"`)
    assertContains(t, resp.Stdout, `"tool":"bash"`)
    assertContains(t, resp.Stdout, `"status":"completed"`)
}
```
