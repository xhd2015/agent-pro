# Scenario

**Feature**: session directory tree created on first append (flat)

```
empty home -> AppendEvent(sessionID, ev) -> sessions/<sessionID>/events.jsonl exists
```

## Preconditions

- Home directory exists but has no `sessions/` subtree before the write.
- First `AppendEvent` must create intermediate directories.
- Path must be `sessions/<session_id>/`, not `sessions/<runner>/<session_id>/`.

## Steps

1. Set `req.Action = "creates_dirs"`.
2. Set `req.SessionID = "sess_new"`.
3. Append one bootstrap event via `Run`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "creates_dirs"
	req.SessionID = "sess_new"
	return nil
}
```