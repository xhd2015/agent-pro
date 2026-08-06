## Expected

- No error.
- Output contains header tokens `NAME`, `SIZE`, `MTIME` (order as in table).
- Output lists `summary.json`, `updates.jsonl`, `signals.json`.

## Errors

- None.

```go
import (
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
	for _, col := range []string{"NAME", "SIZE", "MTIME"} {
		if !strings.Contains(upper, col) {
			t.Fatalf("table missing column %s:\n%s", col, out)
		}
	}
	for _, name := range []string{"summary.json", "updates.jsonl", "signals.json"} {
		assertContains(t, out, name)
	}
}
```
