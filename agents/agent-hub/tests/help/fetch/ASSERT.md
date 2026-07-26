## Expected
- Exit code 0.
- Stdout contains "Usage:", "fetch", describes --consumer-id, --limit, --peek flags.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertExitCode(t, resp, 0)
    assertContains(t, resp.Stdout, "Usage:")
    assertContains(t, resp.Stdout, "fetch")
    assertContains(t, resp.Stdout, "--consumer-id")
    assertContains(t, resp.Stdout, "--limit")
    assertContains(t, resp.Stdout, "--peek")
}
```
