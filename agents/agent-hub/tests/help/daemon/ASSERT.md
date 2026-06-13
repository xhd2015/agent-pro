## Expected
- Exit code 0.
- Stdout contains "Usage:", "daemon", lists subcommands start/stop/status with descriptions.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertExitCode(t, resp, 0)
    assertContains(t, resp.Stdout, "Usage:")
    assertContains(t, resp.Stdout, "daemon")
    assertContains(t, resp.Stdout, "start")
    assertContains(t, resp.Stdout, "stop")
    assertContains(t, resp.Stdout, "status")
}
```
