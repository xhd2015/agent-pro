## Expected
- Exit code 0.
- Stdout contains "Usage:", "replay", describes --consumer-id and --from flags.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertExitCode(t, resp, 0)
    assertContains(t, resp.Stdout, "Usage:")
    assertContains(t, resp.Stdout, "replay")
    assertContains(t, resp.Stdout, "--consumer-id")
    assertContains(t, resp.Stdout, "--from")
}
```
