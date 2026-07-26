## Expected

- `Info` returns an error containing `grok session not found`.
- Error mentions the requested session id.

## Errors

- `grok session not found: 019f283a-eeee-7eee-eeee-eeeeeeeeeeee`

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertError(t, resp)
	if !strings.Contains(resp.Err.Error(), "grok session not found") {
		t.Fatalf("error = %v, want grok session not found", resp.Err)
	}
	if !strings.Contains(resp.Err.Error(), req.SessionID) {
		t.Fatalf("error = %v, want session id %q", resp.Err, req.SessionID)
	}
}
```