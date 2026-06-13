## Expected
- Exit code 0.
- Stdout contains "Usage:", "uninstall", describes --global flag.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertExitCode(t, resp, 0)
    assertContains(t, resp.Stdout, "Usage:")
    assertContains(t, resp.Stdout, "uninstall")
    assertContains(t, resp.Stdout, "--global")
}
```
