---
label: e2e
---

## Expected

- Exit code 0; trailing newline.
- Header includes `UPDATED`.
- Data rows (newest first by absolute time) map session id → relative UPDATED:

| SESSION_ID | UPDATED |
|------------|---------|
| `rel_1h2m` | `1h2m ago` |
| `rel_1h` | `1h ago` |
| `rel_4d5h` | `4d5h ago` |
| `rel_4d` | `4d ago` |
| `rel_90d` | `90d ago` |

- No RFC3339-like `T…Z` timestamps in UPDATED cells.

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
	if !strings.Contains(strings.ToUpper(resp.Stdout), "UPDATED") {
		t.Fatalf("expected UPDATED header:\n%s", resp.Stdout)
	}
	want := map[string]string{
		"rel_1h2m": "1h2m ago",
		"rel_1h":   "1h ago",
		"rel_4d5h": "4d5h ago",
		"rel_4d":   "4d ago",
		"rel_90d":  "90d ago",
	}
	rows := listDataRows(t, resp.Stdout)
	if len(rows) != len(want) {
		t.Fatalf("expected %d data rows, got %d:\n%s", len(want), len(rows), resp.Stdout)
	}
	seen := map[string]string{}
	for _, row := range rows {
		if len(row) < 4 {
			t.Fatalf("expected ≥4 columns (SESSION_ID RUNNER STATUS UPDATED…), got %v\n%s", row, resp.Stdout)
		}
		id := row[0]
		updated := listUpdatedCell(row)
		seen[id] = updated
		if strings.Contains(updated, "T") && (strings.Contains(updated, "Z") || strings.Contains(updated, "+")) {
			t.Fatalf("UPDATED for %s looks absolute RFC3339: %q", id, updated)
		}
	}
	for id, w := range want {
		got, ok := seen[id]
		if !ok {
			t.Fatalf("missing session %q in list:\n%s", id, resp.Stdout)
		}
		if got != w {
			t.Errorf("session %s UPDATED = %q, want %q", id, got, w)
		}
	}
}
```
