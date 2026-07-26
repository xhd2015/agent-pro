# Scenario

**Feature**: info for missing session id returns not-found error

```
# no session JSON for requested id
sessions.Info(dataDir, sessionID) -> error

# error names the missing session id
opencode session not found
```

## Preconditions

- No fixture is written for the requested session id.

## Steps

1. Set `req.SessionID` to a session id that does not exist.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "ses_missing_unknown"
	return nil
}
```