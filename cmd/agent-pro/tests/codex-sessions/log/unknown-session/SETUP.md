# Scenario

**Feature**: log errors when session UUID is not found

```
# no rollout for requested UUID
sessions.Find(codexHome, unknownUUID) -> error

# PrintLog never runs
codex session not found: <id>
```

## Preconditions

- No matching rollout file on disk.

## Steps

1. Set `req.SessionID` to a nonexistent full UUID.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "01900013-4444-7444-8444-444444444444"
	return nil
}
```