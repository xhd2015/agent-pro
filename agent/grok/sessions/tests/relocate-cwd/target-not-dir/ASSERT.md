## Expected

- `RelocateCWD` returns a non-nil error.
- Session directory remains at the original encoded path.
- `summary.json` `info.cwd` is still the old cwd.

## Errors

- Target is not a directory (or equivalent).

```go
import (
	"path/filepath"
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

	oldDir := sessionDirFor(t, req.GrokHome, req.OldCWD, req.SessionID)
	assertDirExists(t, oldDir)
	if got := summaryInfoCWD(t, filepath.Join(oldDir, "summary.json")); got != absPath(t, req.OldCWD) {
		t.Fatalf("info.cwd changed on failure: got %q want %q", got, req.OldCWD)
	}
}
```
