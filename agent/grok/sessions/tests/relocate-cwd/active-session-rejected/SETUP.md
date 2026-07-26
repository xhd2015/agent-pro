# Scenario

**Feature**: active session is rejected without migration

```
seed session + active_sessions.json lists sessionId
-> RelocateCWD(id, existing-target, {GrokHome})
-> error; no move; fields unchanged
```

## Preconditions

- Target directory exists.
- Session is present under encoded old cwd.
- `active_sessions.json` lists the same session id (object form with `sessions` array).

## Steps

1. Create old and new workspaces.
2. Seed session with summary + prompt_context.
3. Seed sqlite marker (must stay untouched).
4. Write `active_sessions.json` containing the session id.
5. Set request fields for asserts.

```go
import (
	"path/filepath"
	"testing"
)

const activeSessionID = "019f283a-eeee-7eee-eeee-eeeeeeeeee05"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	oldWS := filepath.Join(req.TempDir, "ws-old")
	newWS := filepath.Join(req.TempDir, "ws-new")
	mustMkdir(t, oldWS)
	mustMkdir(t, newWS)

	req.OldCWD = absPath(t, oldWS)
	req.TargetDir = absPath(t, newWS)
	req.SessionID = activeSessionID
	req.Active = true
	req.UpdatesMarker = `{"type":"init","marker":"active-no-touch"}` + "\n"
	req.SQLiteMarker = "SQLITE-ACTIVE-MARKER-v1"

	req.SessionDir = writeRelocateSession(t, req.GrokHome, req.SessionID, req.OldCWD, relocateSessionOpts{
		Title:              "active session",
		GitRootEqualsOld:   true,
		WritePromptContext: true,
		UpdatesBody:        req.UpdatesMarker,
	})
	req.SQLitePath = writeSQLiteMarker(t, req.GrokHome, req.SQLiteMarker)
	writeActiveSessions(t, req.GrokHome, req.SessionID)
	return nil
}
```
