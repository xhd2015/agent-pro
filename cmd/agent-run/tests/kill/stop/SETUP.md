# Scenario

**Feature**: `kill` stops a live session; second kill is idempotent

```
startLiveKillFixture
  -> agent-run kill <session-id> -> stopped <session-id>
  -> process dead; registry gone
  -> agent-run kill <session-id> again -> warning not running; exit 0
```

## Preconditions

- Live fixture + session meta for known sessions.
- Mode handle.

## Steps

1. Leaf starts fixture (and optionally pre-kills once for double-kill).
2. Run Handle with kill Args.
3. Assert stdout/stderr contract and liveness side effects.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	req.Mode = "handle"
	return nil
}
```
