# Scenario

**Feature**: enabled + timeout `2m` emits that duration (not the 10m default)

```
BuildFollowUpCommand(ExitOnIdle:true, IdleTimeout:2m, Open, SessionID, Prompt)
  -> --exit-on-idle --idle-timeout=2m (not 10m)
```

## Steps

1. Set `IdleTimeout=2m` with `ExitOnIdle=true`.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.IdleTimeout = 2 * time.Minute
	req.SessionID = "sess-idle-on-2m"
	req.Prompt = "idle on two minutes"
	return nil
}
```
