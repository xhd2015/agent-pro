# Scenario

**Feature**: send queue delivers messages FIFO per session without a blocking trigger send

```
enqueue A --no-wait -> enqueue B --no-wait
  -> session-owned TTY drainer injects A then B (no third default/blocking send)
```

## Preconditions

- Idle stub-tty session with input capture observer attached.
- Session-owned drainer is running on the live serve process.

## Steps

1. `Setup` sets `req.Operation = "fifo"`.
2. `Run` enqueues two messages with `--no-wait` only (A2 — no blocking trigger).
3. `Assert` verifies injection order A before B.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Operation = "fifo"
	req.EnableStubTTY = true
	return nil
}
```
