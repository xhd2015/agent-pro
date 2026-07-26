---
label: e2e
---

## Expected

- Exit code 0.
- Single data row for `no_times` with UPDATED cell exactly `-`.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("expected trailing newline, got %q", resp.Stdout)
	}
	rows := listDataRows(t, resp.Stdout)
	if len(rows) != 1 {
		t.Fatalf("expected 1 data row, got %d:\n%s", len(rows), resp.Stdout)
	}
	if rows[0][0] != "no_times" {
		t.Fatalf("session id = %q, want no_times", rows[0][0])
	}
	updated := listUpdatedCell(rows[0])
	if updated != "-" {
		t.Fatalf("UPDATED = %q, want -", updated)
	}
}
```
