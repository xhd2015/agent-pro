# Scenario

**Feature**: relocate inactive session to existing target directory

```
seed sessions/<encode(ws-old)>/<id>/{summary,prompt_context,updates}
  + session_search.sqlite marker
  + empty active_sessions (or absent)
  + mkdir ws-new
-> RelocateCWD(id, ws-new, {GrokHome})
-> sessions/<encode(ws-new)>/<id>/; info.cwd + working_directory = ws-new
```

## Preconditions

- Session is **not** listed in `active_sessions.json`.
- Target directory exists.
- `summary.json` has `info.cwd` and `git_root_dir` equal to old cwd.
- `prompt_context.json` has `working_directory` = old cwd.
- `updates.jsonl` has a fixed marker body.
- `sessions/session_search.sqlite` has fixed marker bytes.

## Steps

1. Create `ws-old` and `ws-new` under `req.TempDir`.
2. Seed session under encoded old cwd with summary, prompt_context, updates.
3. Seed sqlite marker and empty active sessions list.
4. Set `req.SessionID`, `req.TargetDir`, path markers for asserts.

```go
import (
	"path/filepath"
	"testing"
)

const happySessionID = "019f283a-aaaa-7aaa-aaaa-aaaaaaaaaa01"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	oldWS := filepath.Join(req.TempDir, "ws-old")
	newWS := filepath.Join(req.TempDir, "ws-new")
	mustMkdir(t, oldWS)
	mustMkdir(t, newWS)

	req.OldCWD = absPath(t, oldWS)
	req.TargetDir = absPath(t, newWS)
	req.SessionID = happySessionID
	req.UpdatesMarker = `{"type":"tool_result","current_dir":"` + req.OldCWD + `","marker":"relocate-updates-v1"}` + "\n"
	req.SQLiteMarker = "SQLITE-MARKER-DO-NOT-TOUCH-v1"

	req.SessionDir = writeRelocateSession(t, req.GrokHome, req.SessionID, req.OldCWD, relocateSessionOpts{
		Title:              "happy relocate",
		GitRootEqualsOld:   true,
		WritePromptContext: true,
		UpdatesBody:        req.UpdatesMarker,
	})
	req.SQLitePath = writeSQLiteMarker(t, req.GrokHome, req.SQLiteMarker)
	// Explicit empty active list — session is inactive.
	writeActiveSessions(t, req.GrokHome /* none */)
	return nil
}
```
