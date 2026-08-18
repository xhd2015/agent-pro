# Scenario

**Feature**: enabled + timeout `2s` emits `--idle-timeout=2s`

```
BuildFollowUpCommand(ExitOnIdle:true, IdleTimeout:2s, Open, SessionID, Prompt)
  -> --exit-on-idle --idle-timeout=2s
```

## Steps

1. Set `IdleTimeout=2s` with `ExitOnIdle=true`.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.IdleTimeout = 2 * time.Second
	req.SessionID = "sess-idle-on-2s"
	req.Prompt = "idle on two seconds"
	return nil
}
```
