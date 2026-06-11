## Expected
- The `before_stdout` hook fires.
- stdout is still emitted.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.HookLog, `"event":"BeforeStdout"`)
    assertContains(t, resp.Stdout, `"after hook"`)
}
```

