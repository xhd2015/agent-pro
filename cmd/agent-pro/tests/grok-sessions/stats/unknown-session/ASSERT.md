## Expected

- `Stats` returns an error containing `grok session not found`.
- Error mentions the requested session id.
- `resp.Stats` is nil.

## Errors

- `grok session not found: 019f283b-eeee-7eee-eeee-eeeeeeeeeeee`

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertError(t, resp)
	if !strings.Contains(resp.Err.Error(), "grok session not found") {
		t.Fatalf("error = %v, want grok session not found", resp.Err)
	}
	if !strings.Contains(resp.Err.Error(), req.SessionID) {
		t.Fatalf("error = %v, want session id %q", resp.Err, req.SessionID)
	}
	if resp.Stats != nil {
		t.Fatalf("Stats = %+v, want nil on error", resp.Stats)
	}
}
```
