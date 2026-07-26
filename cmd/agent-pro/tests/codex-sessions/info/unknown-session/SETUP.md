# Scenario

**Feature**: info for missing session id returns not-found error

```
# no rollout file for requested UUID
sessions.Info(codexHome, sessionID) -> error

# error names the missing session id
codex session not found
```

## Preconditions

- No fixture is written for the requested session id.

## Steps

1. Set `req.SessionID` to a UUID that does not exist.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "01900008-eeee-7eee-eeee-eeeeeeeeeeee"
	return nil
}
```