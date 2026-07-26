# Scenario

**Feature**: session metadata CRUD under `sessions/<session_id>/meta.json`

```
CreateSession(sessionID, meta) -> GetSession(sessionID) -> UpdateSessionStatus
ListSessions() returns all sessions (runner is meta only)
GetSession(missing) -> error
CreateSession rejects empty runner and duplicate id
```

## Preconditions

- Session meta includes `runner` (required), `session_id`, `runner_session_id`, `status`, `model`, timestamps.
- Directory layout is flat: `sessions/<session_id>/meta.json`.
- `ListSessions` scans all session dirs under `sessions/` (skips non-session entries such as `.layout`).

## Steps

1. Set `req.Operation = "session"`.
2. Leaf Setup sets `req.Action` and runner/session identifiers.
3. `Run` performs create/get/update, list, or error cases as appropriate.
4. Leaf `Assert` checks meta fields, list contents, paths, or error.

## Context

- `Response.Session` holds the fetched session including meta.
- `Response.Sessions` holds list results from `ListSessions`.
- `Response.Err` is set for missing-session lookups and validation failures.
- `Response.FilesWritten` is populated for create path assertions.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Operation = "session"
	if req.Runner == "" {
		req.Runner = "fake-opencode"
	}
	return nil
}
```