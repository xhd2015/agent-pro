# Scenario

**Feature**: session-owned TTY send-queue drainer delivers without CLI process lifetime

```
# Detached stub-tty serve owns StartSessionDrainer for terminalSessionID
agent-run run --agent-runner stub-tty --keep-tty ... -> ServeSession -> session drainer loop

# --no-wait CLI only enqueues then exits; delivery is independent of CLI
agent-run send <session-id> "msg" --no-wait -> Enqueue -> stdout msg_N -> exit 0
  -> (CLI gone) session drainer drainStep -> WriteInput / scrollback
```

## Preconditions

- Background stub-tty remains alive for the duration of each leaf (session-owned consumer).
- No leaf under this branch issues a blocking (default) send solely to "wake" a CLI drainer.
- Delivery correctness must hold with only `--no-wait` CLI invocations.

## Steps

1. `Setup` sets `req.Operation = "tty-drainer"`.
2. Leaf chooses idle (A1) or busy-then-idle (A3) scenario and message text.
3. `Run` starts stub session, executes `--no-wait` only, waits for inject / status.
4. `Assert` checks fast CLI return (where applicable), inject, status delivered, queue empty.

```go
func Setup(t *testing.T, req *Request) error {
	req.Operation = "tty-drainer"
	req.EnableStubTTY = true
	return nil
}
```
