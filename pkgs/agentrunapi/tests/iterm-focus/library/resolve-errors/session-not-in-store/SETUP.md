# Scenario

**Feature**: session id not present in store

```
GetSession(unknown) -> FocusSession error; no focus
```

## Steps

1. Request a session id that was never created; store home is empty of sessions.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "sess-does-not-exist"
	req.SeedSession = false
	req.SeedRegistry = false
	return nil
}
```
