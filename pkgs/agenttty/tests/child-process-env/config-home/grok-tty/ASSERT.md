## Expected

- `Set` contains `GROK_HOME=<configHome>`.
- `Set` does not contain `CODEX_HOME`.
- `Unset` is empty.

## Errors

- Emitting CODEX_HOME for grok-tty; missing GROK_HOME.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertBuildOK(t, resp, err)
	assertSetExact(t, resp.Set, "GROK_HOME", req.ConfigHome)
	assertSetAbsent(t, resp.Set, "CODEX_HOME")
	assertUnsetEmpty(t, resp.Unset)
}
```
