## Expected
- Exit code 0.
- Stdout contains "Usage:", "integration", lists "opencode" and "status" subcommands.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertExitCode(t, resp, 0)
    assertContains(t, resp.Stdout, "Usage:")
    assertContains(t, resp.Stdout, "integration")
    assertContains(t, resp.Stdout, "opencode")
    assertContains(t, resp.Stdout, "status")
}
```
