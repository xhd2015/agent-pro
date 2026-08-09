## Expected

- Set contains `FOO=bar`.
- Unset is empty.
- No force-color keys.

## Errors

- Missing FOO; inventing color Unset when Color=false.

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
	assertSetExact(t, resp.Set, "FOO", "bar")
	assertUnsetEmpty(t, resp.Unset)
	assertSetAbsent(t, resp.Set, "FORCE_COLOR")
}
```
