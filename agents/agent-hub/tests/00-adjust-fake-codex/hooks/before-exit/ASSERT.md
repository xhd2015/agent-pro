## Expected
- The `before_exit` hook fires before process completion.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.HookLog, `"event":"BeforeExit"`)
}
```

