## Expected
- Exit code 0.
- Stdout contains "Usage:", "pop", describes --runner, --session-id flags.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertExitCode(t, resp, 0)
    assertContains(t, resp.Stdout, "Usage:")
    assertContains(t, resp.Stdout, "pop")
    assertContains(t, resp.Stdout, "--runner")
    assertContains(t, resp.Stdout, "--session-id")
}
```
