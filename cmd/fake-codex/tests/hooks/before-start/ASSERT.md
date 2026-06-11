## Expected
- The `before_start` hook fires once.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.HookLog, `"event":"SessionStart"`)
}
```

