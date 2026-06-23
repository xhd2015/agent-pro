## Expected

- `Find` returns error containing `codex session not found`.
- Error mentions the requested session id.
- Log output is empty.

## Errors

- `codex session not found: 01900013-4444-7444-8444-444444444444`

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertError(t, resp)
	if !strings.Contains(resp.Err.Error(), "codex session not found") {
		t.Fatalf("error = %v, want codex session not found", resp.Err)
	}
	if !strings.Contains(resp.Err.Error(), req.SessionID) {
		t.Fatalf("error = %v, want session id %q", resp.Err, req.SessionID)
	}
	if strings.TrimSpace(resp.Output) != "" {
		t.Fatalf("expected empty output on error, got:\n%s", resp.Output)
	}
}
```