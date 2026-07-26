## Expected
- Exit code 0.
- Stdout contains "Usage:", "session message", lists send/list/pop subcommands.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertExitCode(t, resp, 0)
    assertContains(t, resp.Stdout, "Usage:")
    assertContains(t, resp.Stdout, "message")
    assertContains(t, resp.Stdout, "send")
    assertContains(t, resp.Stdout, "list")
    assertContains(t, resp.Stdout, "pop")
}
```
