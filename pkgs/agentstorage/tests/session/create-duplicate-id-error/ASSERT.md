## Expected

- Second `CreateSession` returns a non-nil error.
- Flat session directory still exists after the failed create.
- No nested `sessions/<runner>/<id>/` layout.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	// Run returns (resp, nil) with resp.Err set for the second CreateSession.
	assertError(t, resp.Err)
	flatMeta := filepath.Join(resp.ResolvedHome, "sessions", req.SessionID, "meta.json")
	if _, err := os.Stat(flatMeta); err != nil {
		t.Fatalf("expected first session meta to remain at %q: %v", flatMeta, err)
	}
	nestedDir := filepath.Join(resp.ResolvedHome, "sessions", req.Runner, req.SessionID)
	if st, err := os.Stat(nestedDir); err == nil && st.IsDir() {
		t.Fatalf("unexpected nested runner session dir %q", nestedDir)
	}
}
```
