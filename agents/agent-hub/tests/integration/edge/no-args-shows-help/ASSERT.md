## Expected
- Exit code 0.
- Stdout contains help text with "Usage:", "status", "install", "uninstall", "enable", "disable".

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertExitCode(t, resp, 0)
    assertContains(t, resp.Stdout, "Usage:")
    assertContains(t, resp.Stdout, "status")
    assertContains(t, resp.Stdout, "install")
    assertContains(t, resp.Stdout, "uninstall")
    assertContains(t, resp.Stdout, "enable")
    assertContains(t, resp.Stdout, "disable")
}
```
