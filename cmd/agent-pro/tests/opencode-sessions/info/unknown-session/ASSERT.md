## Expected

- `Info` returns an error containing `opencode session not found`.
- Error mentions the requested session id.

## Errors

- `opencode session not found: ses_missing_unknown`

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertError(t, resp)
	if !strings.Contains(resp.Err.Error(), "opencode session not found") {
		t.Fatalf("error = %v, want opencode session not found", resp.Err)
	}
	if !strings.Contains(resp.Err.Error(), req.SessionID) {
		t.Fatalf("error = %v, want session id %q", resp.Err, req.SessionID)
	}
}
```