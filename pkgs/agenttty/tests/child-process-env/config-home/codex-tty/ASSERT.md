## Expected

- `Set` contains `CODEX_HOME=<configHome>`.
- `Set` does not contain `GROK_HOME`.
- `Unset` is empty.

## Errors

- Emitting GROK_HOME for codex-tty; missing CODEX_HOME.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertBuildOK(t, resp, err)
	assertSetExact(t, resp.Set, "CODEX_HOME", req.ConfigHome)
	assertSetAbsent(t, resp.Set, "GROK_HOME")
	assertUnsetEmpty(t, resp.Unset)
}
```
