## Expected

- Exit code 0.
- Exactly 10 data rows (excluding optional header/footer).
- Newest-first: first data row is `sess_14`, then `sess_13`, … down to `sess_05`.
- No compound `runner/id` in session id column.

```go
import (
	"fmt"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("expected trailing newline, got %q", resp.Stdout)
	}
	rows := listDataRows(t, resp.Stdout)
	if len(rows) != 10 {
		t.Fatalf("expected 10 data rows (default limit), got %d:\n%s", len(rows), resp.Stdout)
	}
	for i := 0; i < 10; i++ {
		wantID := fmt.Sprintf("sess_%02d", 14-i)
		gotID := rows[i][0]
		if gotID != wantID {
			t.Fatalf("row %d: session_id=%q want %q\nstdout:\n%s", i, gotID, wantID, resp.Stdout)
		}
		if strings.Contains(gotID, "/") {
			t.Fatalf("compound session ref %q", gotID)
		}
	}
}
```
