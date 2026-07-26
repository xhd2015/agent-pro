## Expected

- `RelocateCWD` returns a non-nil error (active session rejected).
- Result is nil.
- Old session directory still exists; new encoded target path does **not**.
- `summary.json` `info.cwd` remains old cwd.
- `prompt_context.json` `working_directory` remains old cwd.
- `updates.jsonl` content unchanged at the old path.
- `session_search.sqlite` marker bytes unchanged.

## Errors

- Active / in-use session (or equivalent wording).

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run harness error: %v", err)
	}
	assertError(t, resp)
	if resp.Result != nil {
		t.Fatalf("expected nil Result on error, got %+v", resp.Result)
	}
	msg := strings.ToLower(resp.Err.Error())
	if !strings.Contains(msg, "active") {
		t.Fatalf("error %q should mention active session", resp.Err)
	}

	oldDir := sessionDirFor(t, req.GrokHome, req.OldCWD, req.SessionID)
	newDir := sessionDirFor(t, req.GrokHome, req.TargetDir, req.SessionID)
	assertDirExists(t, oldDir)
	assertPathMissing(t, newDir)

	wantOld := absPath(t, req.OldCWD)
	if got := summaryInfoCWD(t, filepath.Join(oldDir, "summary.json")); got != wantOld {
		t.Fatalf("info.cwd changed while active: got %q want %q", got, wantOld)
	}
	if got := promptWorkingDirectory(t, filepath.Join(oldDir, "prompt_context.json")); got != wantOld {
		t.Fatalf("working_directory changed while active: got %q want %q", got, wantOld)
	}
	assertFileEquals(t, filepath.Join(oldDir, "updates.jsonl"), req.UpdatesMarker)
	assertFileEquals(t, req.SQLitePath, req.SQLiteMarker)
}
```
