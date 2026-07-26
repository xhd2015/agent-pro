# Scenario

**Feature**: `agent-run send cancel` removes pending queue entries

```
send cancel <session-id> msg_N -> flock -> remove pending line (silent success)
```

## Preconditions

- Stub-tty session for resolve-by-terminal-id.
- Pending message enqueued via `--no-wait` where applicable.

## Steps

1. `Setup` sets `req.Operation = "cancel"`.
2. Leaf `Setup` sets `req.Action` and message text.
3. `Run` enqueues (when needed) then runs `send cancel`.
4. `Assert` checks exit code, stderr, injection absence, or delivered-state failure.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Operation = "cancel"
	req.EnableStubTTY = true
	return nil
}
```