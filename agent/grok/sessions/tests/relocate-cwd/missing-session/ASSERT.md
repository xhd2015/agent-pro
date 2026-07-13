## Expected

- `RelocateCWD` returns a non-nil error.
- Error message mentions the session id.
- Error indicates the session was not found (substring `not found` or equivalent).
- Result is nil.

## Errors

- Session not found for the requested id.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run harness error: %v", err)
	}
	assertError(t, resp)
	if resp.Result != nil {
		t.Fatalf("expected nil Result on error, got %+v", resp.Result)
	}
	msg := resp.Err.Error()
	if !strings.Contains(msg, req.SessionID) {
		t.Fatalf("error %q should mention session id %q", msg, req.SessionID)
	}
	lower := strings.ToLower(msg)
	if !strings.Contains(lower, "not found") && !strings.Contains(lower, "unknown") {
		t.Fatalf("error %q should indicate session not found", msg)
	}
}
```
