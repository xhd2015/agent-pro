## Expected
- The tool event keeps the shape parsed by the existing opencode runner.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"tool_use"`)
    assertContains(t, resp.Stdout, `"tool":"bash"`)
    assertContains(t, resp.Stdout, `"callID":"call_1"`)
    assertContains(t, resp.Stdout, `"status":"completed"`)
}
```

