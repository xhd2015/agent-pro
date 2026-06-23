## Expected

- `List` returns zero sessions without error.
- `FormatListTable` output is exactly `No sessions found`.

## Errors

- None.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if len(resp.Sessions) != 0 {
		t.Fatalf("len(sessions) = %d, want 0", len(resp.Sessions))
	}
	if strings.TrimSpace(resp.Output) != "No sessions found" {
		t.Fatalf("output = %q, want %q", resp.Output, "No sessions found")
	}
}
```