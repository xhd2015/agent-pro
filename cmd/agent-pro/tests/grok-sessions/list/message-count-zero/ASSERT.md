## Expected

- `FormatListTable` output contains column header `MSGS`.
- Output includes `0` for the empty-title session.
- Session UUID appears in the table row.

## Errors

- None.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Output, "MSGS")
	assertContains(t, resp.Output, "01900005-bbbb-7bbb-bbbb-bbbbbbbbbbbb")

	lines := strings.Split(resp.Output, "\n")
	if len(lines) < 2 {
		t.Fatalf("expected header and data row, got:\n%s", resp.Output)
	}
	dataRow := lines[1]
	if !strings.Contains(dataRow, " 0 ") && !strings.HasSuffix(strings.TrimSpace(dataRow), " 0") {
		t.Fatalf("data row missing MSGS value 0:\n%s", dataRow)
	}
}
```