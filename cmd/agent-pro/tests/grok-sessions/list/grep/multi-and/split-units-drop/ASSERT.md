## Expected

- No matching sessions.
- Output is `No sessions found` (or equivalent empty list message).

## Errors

- None.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if len(resp.Sessions) != 0 {
		t.Fatalf("len(sessions) = %d, want 0", len(resp.Sessions))
	}
	if !strings.Contains(resp.Output, "No sessions found") {
		t.Fatalf("want No sessions found, got:\n%s", resp.Output)
	}
}
```
