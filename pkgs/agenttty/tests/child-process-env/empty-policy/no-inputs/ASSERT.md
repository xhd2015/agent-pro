## Expected

- `err` is nil.
- `resp.Set` is empty.
- `resp.Unset` is empty.
- No FORCE_COLOR / CLICOLOR / TERM / CODEX_HOME / GROK_HOME / PATH in Set.

## Errors

- Accidental color or home injection on empty inputs.

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
	assertSetEmpty(t, resp.Set)
	assertUnsetEmpty(t, resp.Unset)
	for _, key := range []string{"FORCE_COLOR", "CLICOLOR", "CLICOLOR_FORCE", "TERM", "NO_COLOR", "CODEX_HOME", "GROK_HOME", "PATH"} {
		assertSetAbsent(t, resp.Set, key)
	}
}
```
