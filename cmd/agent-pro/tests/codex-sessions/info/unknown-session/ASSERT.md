## Expected

- `Info` returns an error containing `codex session not found`.
- Error mentions the requested session id.

## Errors

- `codex session not found: 01900008-eeee-7eee-eeee-eeeeeeeeeeee`

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
}
```