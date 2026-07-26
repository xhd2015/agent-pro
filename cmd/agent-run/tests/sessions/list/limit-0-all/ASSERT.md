---
label: e2e
---

## Expected

- Exit code 0.
- Exactly 15 data rows, newest first (`sess_14` … `sess_00`).

```go
import (
	"fmt"
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
	if len(rows) != 15 {
		t.Fatalf("expected 15 data rows (--limit 0 = all), got %d:\n%s", len(rows), resp.Stdout)
	}
	for i := 0; i < 15; i++ {
		wantID := fmt.Sprintf("sess_%02d", 14-i)
		if rows[i][0] != wantID {
			t.Fatalf("row %d: got %q want %q\n%s", i, rows[i][0], wantID, resp.Stdout)
		}
	}
}
```
