## Expected

- Success: session relocated under `opts.GrokHome` (custom home).
- New session path is under custom home only.
- Old path under custom home is gone.
- Decoy home marker file and sqlite content are unchanged.
- Result paths are under the custom home prefix.
- Custom home sqlite marker unchanged.

## Errors

- None.

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run harness error: %v", err)
	}
	assertNoError(t, resp)
	if resp.Result == nil {
		t.Fatal("expected non-nil RelocateCWDResult")
	}

	customHome := req.OptsGrokHome
	wantNew := absPath(t, req.TargetDir)
	wantOld := absPath(t, req.OldCWD)
	wantOldDir := sessionDirFor(t, customHome, wantOld, req.SessionID)
	wantNewDir := sessionDirFor(t, customHome, wantNew, req.SessionID)

	assertPathMissing(t, wantOldDir)
	assertDirExists(t, wantNewDir)

	if !strings.HasPrefix(filepath.Clean(resp.Result.NewSessionDir), filepath.Clean(customHome)) {
		t.Fatalf("NewSessionDir %q not under custom home %q", resp.Result.NewSessionDir, customHome)
	}
	if !strings.HasPrefix(filepath.Clean(resp.Result.OldSessionDir), filepath.Clean(customHome)) {
		t.Fatalf("OldSessionDir %q not under custom home %q", resp.Result.OldSessionDir, customHome)
	}
	if resp.Result.NewCWD != wantNew {
		t.Fatalf("NewCWD = %q, want %q", resp.Result.NewCWD, wantNew)
	}
	if got := summaryInfoCWD(t, filepath.Join(wantNewDir, "summary.json")); got != wantNew {
		t.Fatalf("info.cwd = %q, want %q", got, wantNew)
	}
	if got := promptWorkingDirectory(t, filepath.Join(wantNewDir, "prompt_context.json")); got != wantNew {
		t.Fatalf("working_directory = %q, want %q", got, wantNew)
	}

	// Decoy home must not be modified.
	assertFileEquals(t, filepath.Join(req.DecoyGrokHome, "sessions", "decoy-marker.txt"), "decoy-stay\n")
	assertFileEquals(t, filepath.Join(req.DecoyGrokHome, "sessions", "session_search.sqlite"), "DECOY-SQLITE\n")
	// No session for this id under decoy.
	assertPathMissing(t, sessionDirFor(t, req.DecoyGrokHome, wantNew, req.SessionID))
	assertPathMissing(t, sessionDirFor(t, req.DecoyGrokHome, wantOld, req.SessionID))

	assertFileEquals(t, req.SQLitePath, req.SQLiteMarker)
}
```
