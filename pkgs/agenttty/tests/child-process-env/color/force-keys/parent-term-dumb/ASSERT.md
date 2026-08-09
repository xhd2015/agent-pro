## Expected

- Unset contains `NO_COLOR`.
- Set contains `FORCE_COLOR=1`, `CLICOLOR=1`, `CLICOLOR_FORCE=1`.
- Set contains `TERM=xterm-256color` (dumb rewritten).

## Errors

- Leaving TERM=dumb; missing force keys; empty Unset.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	assertBuildOK(t, resp, err)
	assertColorForceKeys(t, resp.Set, resp.Unset)
	assertSetExact(t, resp.Set, "TERM", "xterm-256color")
}
```
