## Expected

- No error.
- One session.
- FormatListTableWithHits header has SESSION ID then KIND then LAST ACTIVE.

## Errors

- None.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertSessionIDs(t, resp.Sessions, idMain)
	if strings.TrimSpace(resp.Output) == "" {
		t.Fatal("expected hits table output, got empty")
	}
	assertHeaderKINDColumn(t, resp.Output)
}
```
