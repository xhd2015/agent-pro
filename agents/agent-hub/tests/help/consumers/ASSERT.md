## Expected
- Exit code 0.
- Stdout contains "Usage:", "consumers", describes listing consumers.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertExitCode(t, resp, 0)
    assertContains(t, resp.Stdout, "Usage:")
    assertContains(t, resp.Stdout, "consumers")
}
```
