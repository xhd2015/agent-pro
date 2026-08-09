## Expected

- Unset contains `NO_COLOR`.
- Force color keys present with value `1`.
- `TERM=xterm-256color` in Set.

## Errors

- Missing TERM assignment when parentTERM empty.

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
