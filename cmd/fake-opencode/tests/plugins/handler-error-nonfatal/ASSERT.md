## Expected
- Exit code 0 (plugin errors are non-fatal).
- Stderr contains "plugin handler failed".

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stderr, "plugin handler failed")
}
```
