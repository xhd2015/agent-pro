# Scenario

**Feature**: disabled + timeout 0 omits both idle flags

```
BuildFollowUpCommand(ExitOnIdle:false, IdleTimeout:0, Open, SessionID, Prompt)
  -> no --exit-on-idle / --idle-timeout tokens
```

## Steps

1. Set `IdleTimeout=0` with open profile.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.IdleTimeout = 0
	req.SessionID = "sess-idle-off-zero"
	req.Prompt = "idle off zero"
	return nil
}
```
