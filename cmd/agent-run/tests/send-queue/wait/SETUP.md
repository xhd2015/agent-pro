# Scenario

**Feature**: send wait modes control CLI blocking and timeout behavior (REG)

```
default -> enqueue + poll until msg_N absent (session TTY drainer delivers)
--no-wait -> return immediately after enqueue (return-before-inject on busy)
--max-wait DURATION -> wall-clock timeout removes only caller message
```

## Preconditions

- Stub-tty scenario controls busy vs idle writable state.
- Queue file inspectable under `send-queue/<runner>/<session-id>.jsonl`.
- Eventual `--no-wait` delivery without CLI lifetime is covered under `tty-drainer/`.

## Steps

1. `Setup` sets `req.Operation = "wait"`.
2. Leaf selects busy, busy-then-idle, or permanently busy scenario via `req.Action`.
3. `Run` executes send with appropriate flags and records timing / queue state.
4. `Assert` checks delivery, timeout stderr, or fast return semantics.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Operation = "wait"
	req.EnableStubTTY = true
	return nil
}
```