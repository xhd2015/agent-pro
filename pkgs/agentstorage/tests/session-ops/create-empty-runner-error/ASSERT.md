## Expected

- `CreateSession` returns a non-nil error.
- No `sessions/<session_id>/` directory is created.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertError(t, resp.Err)
	dir := filepath.Join(resp.ResolvedHome, "sessions", req.SessionID)
	if st, err := os.Stat(dir); err == nil && st.IsDir() {
		t.Fatalf("expected no session dir after empty-runner create, found %q", dir)
	}
}
```
