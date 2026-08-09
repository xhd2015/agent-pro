## Expected

- Unset contains `NO_COLOR`.
- Set does **not** contain `NO_COLOR=…`.
- Force color keys present (`FORCE_COLOR`, `CLICOLOR`, `CLICOLOR_FORCE` = 1).

## Errors

- Leaving `NO_COLOR=1` in Set (would reintroduce after unset at merge).

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
	assertUnsetHas(t, resp.Unset, "NO_COLOR")
	assertSetAbsent(t, resp.Set, "NO_COLOR")
	assertSetExact(t, resp.Set, "FORCE_COLOR", "1")
	assertSetExact(t, resp.Set, "CLICOLOR", "1")
	assertSetExact(t, resp.Set, "CLICOLOR_FORCE", "1")
}
```
