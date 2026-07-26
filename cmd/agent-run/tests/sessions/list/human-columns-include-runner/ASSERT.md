---
label: e2e
---

## Expected

- Exit code 0.
- Header line includes `SESSION_ID`, `RUNNER`, and `STATUS` (UPDATED recommended).
- Data rows include bare ids `demo_b` then `demo_a` (newest first).
- Runner values `fake-opencode` and `fake-codex` appear as columns, not as `runner/id` prefixes.

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
	upper := strings.ToUpper(resp.Stdout)
	for _, col := range []string{"SESSION_ID", "RUNNER", "STATUS"} {
		if !strings.Contains(upper, col) {
			t.Fatalf("expected header column %s in stdout:\n%s", col, resp.Stdout)
		}
	}
	rows := listDataRows(t, resp.Stdout)
	if len(rows) != 2 {
		t.Fatalf("expected 2 data rows, got %d:\n%s", len(rows), resp.Stdout)
	}
	// newest first: demo_b then demo_a
	if rows[0][0] != "demo_b" || rows[1][0] != "demo_a" {
		t.Fatalf("unexpected order: %v then %v\n%s", rows[0], rows[1], resp.Stdout)
	}
	// find runner field: typically index 1 when columns are SESSION_ID RUNNER STATUS UPDATED
	joined0 := strings.Join(rows[0], " ")
	joined1 := strings.Join(rows[1], " ")
	if !strings.Contains(joined0, "fake-opencode") {
		t.Fatalf("expected runner fake-opencode on demo_b row: %q", joined0)
	}
	if !strings.Contains(joined1, "fake-codex") {
		t.Fatalf("expected runner fake-codex on demo_a row: %q", joined1)
	}
	if strings.Contains(rows[0][0], "/") || strings.Contains(rows[1][0], "/") {
		t.Fatalf("session id column must be bare id, got %q / %q", rows[0][0], rows[1][0])
	}
	// old format was runner/id as first token
	if strings.Contains(resp.Stdout, "fake-codex/demo_a") || strings.Contains(resp.Stdout, "fake-opencode/demo_b") {
		t.Fatalf("stdout must not use runner/id compound refs:\n%s", resp.Stdout)
	}
}
```
