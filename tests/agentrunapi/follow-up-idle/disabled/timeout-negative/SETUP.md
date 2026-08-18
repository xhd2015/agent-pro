# Scenario

**Feature**: disabled + negative timeout is still a silent no-op

```
NormalizeIdle(false, -1s) -> enabled=false, no error
BuildFollowUpCommand -> no idle-exit tokens
```

## Steps

1. Set `IdleTimeout=-1s` with `ExitOnIdle=false`.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.IdleTimeout = -1 * time.Second
	req.SessionID = "sess-idle-off-neg"
	req.Prompt = "idle off negative"
	return nil
}
```
