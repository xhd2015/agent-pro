## Expected

- `RelocateCWD` returns nil error and non-nil result.
- Result `OldCWD` / `NewCWD` match pre-migration abs old path and target.
- Result `OldSessionDir` / `NewSessionDir` match encoded old/new layout paths.
- Old session directory is gone; new session directory exists.
- `summary.json` `info.cwd` equals target abs path.
- `summary.json` `git_root_dir` equals target abs path (was equal to old cwd).
- `prompt_context.json` `working_directory` equals target abs path.
- `updates.jsonl` content at the **new** session path equals the seeded marker
  (content preserved with the move; not bulk-rewritten to erase marker).
- `sessions/session_search.sqlite` still exists with identical marker bytes
  (not modified).

## Errors

- None.

```go
import (
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run harness error: %v", err)
	}
	assertNoError(t, resp)
	if resp.Result == nil {
		t.Fatal("expected non-nil RelocateCWDResult")
	}

	wantNew := absPath(t, req.TargetDir)
	wantOld := absPath(t, req.OldCWD)
	wantOldDir := sessionDirFor(t, req.GrokHome, wantOld, req.SessionID)
	wantNewDir := sessionDirFor(t, req.GrokHome, wantNew, req.SessionID)

	if resp.Result.OldCWD != wantOld {
		t.Fatalf("OldCWD = %q, want %q", resp.Result.OldCWD, wantOld)
	}
	if resp.Result.NewCWD != wantNew {
		t.Fatalf("NewCWD = %q, want %q", resp.Result.NewCWD, wantNew)
	}
	if filepath.Clean(resp.Result.OldSessionDir) != filepath.Clean(wantOldDir) {
		t.Fatalf("OldSessionDir = %q, want %q", resp.Result.OldSessionDir, wantOldDir)
	}
	if filepath.Clean(resp.Result.NewSessionDir) != filepath.Clean(wantNewDir) {
		t.Fatalf("NewSessionDir = %q, want %q", resp.Result.NewSessionDir, wantNewDir)
	}

	assertPathMissing(t, wantOldDir)
	assertDirExists(t, wantNewDir)

	summaryPath := filepath.Join(wantNewDir, "summary.json")
	if got := summaryInfoCWD(t, summaryPath); got != wantNew {
		t.Fatalf("info.cwd = %q, want %q", got, wantNew)
	}
	sum := readJSONMap(t, summaryPath)
	if gitRoot, _ := sum["git_root_dir"].(string); gitRoot != wantNew {
		t.Fatalf("git_root_dir = %q, want %q", gitRoot, wantNew)
	}

	promptPath := filepath.Join(wantNewDir, "prompt_context.json")
	if got := promptWorkingDirectory(t, promptPath); got != wantNew {
		t.Fatalf("working_directory = %q, want %q", got, wantNew)
	}

	updatesPath := filepath.Join(wantNewDir, "updates.jsonl")
	assertFileEquals(t, updatesPath, req.UpdatesMarker)

	assertFileEquals(t, req.SQLitePath, req.SQLiteMarker)
}
```
