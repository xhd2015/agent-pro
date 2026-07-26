# Scenario

**Feature**: successful enqueue prints session-local message id to stdout

```
agent-run send <session-id> "msg" -> agentsend.Enqueue -> stdout msg_N\n
```

## Preconditions

- Background stub-tty session with idle writable prompt (keep-alive scenario).
- Registry entry reachable for the terminal session id.

## Steps

1. `Setup` enables stub-tty and sets `req.Operation = "enqueue"`.
2. Leaf `Setup` sets `req.Action` and message text.
3. `Run` starts stub session (if not already) and executes send CLI.
4. `Assert` checks stdout id line and exit code.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Operation = "enqueue"
	req.EnableStubTTY = true
	return nil
}
```