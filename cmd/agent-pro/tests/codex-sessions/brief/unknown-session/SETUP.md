# Scenario

**Feature**: brief errors when session UUID is not found

```
# empty Codex home, no matching rollout
sessions.Brief(codexHome, unknownUUID) -> error

# error message names the missing session
codex session not found: <id>
```

## Preconditions

- No rollout file exists for the requested session id.
- Full UUID is required (no prefix matching).

## Steps

1. Set `req.SessionID` to a UUID that does not exist on disk.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "01900008-eeee-7eee-eeee-eeeeeeeeeeee"
	return nil
}
```