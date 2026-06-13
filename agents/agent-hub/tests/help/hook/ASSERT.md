## Expected
- Exit code 0.
- Stdout contains "Usage:", "hook", "notify", describes --runner and --event flags.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertExitCode(t, resp, 0)
    assertContains(t, resp.Stdout, "Usage:")
    assertContains(t, resp.Stdout, "hook")
    assertContains(t, resp.Stdout, "notify")
    assertContains(t, resp.Stdout, "--runner")
    assertContains(t, resp.Stdout, "--event")
}
```
