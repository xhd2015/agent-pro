## Expected
- Exit code 0.
- Stdout contains "Usage:", "session", lists show and message subcommands.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertExitCode(t, resp, 0)
    assertContains(t, resp.Stdout, "Usage:")
    assertContains(t, resp.Stdout, "session")
    assertContains(t, resp.Stdout, "show")
    assertContains(t, resp.Stdout, "message")
}
```
