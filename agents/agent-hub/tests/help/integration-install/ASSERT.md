## Expected
- Exit code 0.
- Stdout contains "Usage:", "integration install", describes <runner> arg and --global flag.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertExitCode(t, resp, 0)
    assertContains(t, resp.Stdout, "Usage:")
    assertContains(t, resp.Stdout, "install")
    assertContains(t, resp.Stdout, "--global")
}
```
