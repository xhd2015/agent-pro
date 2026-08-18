# Scenario

**Feature**: disabled + timeout `2s` still omits both idle flags

```
BuildFollowUpCommand(ExitOnIdle:false, IdleTimeout:2s, Open, SessionID, Prompt)
  -> no --exit-on-idle / --idle-timeout tokens (timeout ignored)
```

## Steps

1. Set `IdleTimeout=2s` with `ExitOnIdle=false`.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.IdleTimeout = 2 * time.Second
	req.SessionID = "sess-idle-off-2s"
	req.Prompt = "idle off two seconds"
	return nil
}
```
