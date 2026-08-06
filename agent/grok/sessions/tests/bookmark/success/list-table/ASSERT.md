## Expected

- No error; Output non-empty.
- Output contains header tokens RUNNER, SESSION (or SESSION ID), MSGS, TITLE
  (case-insensitive check OK).
- Output contains `grok`, session id, title, and message count `42`.
- Tags `backup` / `migration` appear when table includes a tags column.

## Errors

- None.

```go
import (
	"strconv"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	out := resp.Output
	if out == "" {
		t.Fatal("table output empty")
	}
	upper := strings.ToUpper(out)
	for _, col := range []string{"RUNNER", "SESSION", "MSGS", "TITLE"} {
		if !strings.Contains(upper, col) {
			t.Fatalf("table missing column %s:\n%s", col, out)
		}
	}
	assertContains(t, out, "grok")
	assertContains(t, out, req.SessionID)
	assertContains(t, out, req.Title)
	assertContains(t, out, strconv.Itoa(req.NumChatMessages))
	// tags optional column but preseeded values should surface when present
	if strings.Contains(upper, "TAG") {
		assertContains(t, out, "backup")
		assertContains(t, out, "migration")
	}
}
```
