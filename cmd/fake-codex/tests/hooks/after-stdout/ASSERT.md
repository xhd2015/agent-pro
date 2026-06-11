## Expected
- The `after_stdout` hook fires.
- stdout is emitted.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"before hook"`)
    assertContains(t, resp.HookLog, `"event":"AfterStdout"`)
}
```

