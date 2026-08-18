# Scenario

**Feature**: enabled + negative timeout is an API error

```
NormalizeIdle(true, -1s) -> error
BuildFollowUpCommand(ExitOnIdle:true, IdleTimeout:-1s) -> API error (normalize on emit path)
```

## Steps

1. Set `IdleTimeout=-1s` with `ExitOnIdle=true`.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.IdleTimeout = -1 * time.Second
	req.SessionID = "sess-idle-on-neg"
	req.Prompt = "idle on negative"
	return nil
}
```
