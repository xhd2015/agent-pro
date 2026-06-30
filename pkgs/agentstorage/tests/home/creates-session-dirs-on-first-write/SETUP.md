# Scenario

**Feature**: session directory tree created on first append

```
empty home -> AppendEvent(runner, sessionID, ev) -> sessions/<runner>/<sessionID>/events.jsonl exists
```

## Preconditions

- Home directory exists but has no `sessions/` subtree before the write.
- First `AppendEvent` must create intermediate directories.

## Steps

1. Set `req.Action = "creates_dirs"`.
2. Set `req.SessionID = "sess_new"`.
3. Append one bootstrap event via `Run`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Action = "creates_dirs"
	req.SessionID = "sess_new"
	return nil
}
```