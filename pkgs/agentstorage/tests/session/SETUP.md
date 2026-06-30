# Scenario

**Feature**: session metadata CRUD under `sessions/<runner>/<session_id>/meta.json`

```
CreateSession(meta) -> GetSession -> UpdateSessionStatus -> GetSession
ListSessions(runner) filters by runner prefix
GetSession(missing) -> error
```

## Preconditions

- Session meta includes `runner`, `session_id`, `runner_session_id`, `status`, `model`, timestamps.
- `ListSessions` scans only `sessions/<runner>/` subtree.

## Steps

1. Set `req.Operation = "session"`.
2. Leaf Setup sets `req.Action` and runner/session identifiers.
3. `Run` performs create/get/update, list, or missing-get as appropriate.
4. Leaf `Assert` checks meta fields, list contents, or error.

## Context

- `Response.Session` holds the fetched session including meta.
- `Response.Sessions` holds list results from `ListSessions`.
- `Response.Err` is set for missing-session lookups.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Operation = "session"
	if req.Runner == "" {
		req.Runner = "fake-opencode"
	}
	return nil
}
```