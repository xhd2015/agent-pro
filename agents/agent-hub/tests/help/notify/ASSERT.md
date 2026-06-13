## Expected
- Exit code 0.
- Stdout contains "Usage:", "notify", describes --json and --file flags.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertExitCode(t, resp, 0)
    assertContains(t, resp.Stdout, "Usage:")
    assertContains(t, resp.Stdout, "notify")
    assertContains(t, resp.Stdout, "--json")
    assertContains(t, resp.Stdout, "--file")
}
```
