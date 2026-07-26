---
label: e2e
---

## Expected

- Exit code 0.
- Exactly 3 data rows: `sess_14`, `sess_13`, `sess_12` in that order.

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
	if len(rows) != 3 {
		t.Fatalf("expected 3 data rows, got %d:\n%s", len(rows), resp.Stdout)
	}
	for i := 0; i < 3; i++ {
		wantID := fmt.Sprintf("sess_%02d", 14-i)
		if rows[i][0] != wantID {
			t.Fatalf("row %d: got %q want %q\n%s", i, rows[i][0], wantID, resp.Stdout)
		}
	}
}
```
