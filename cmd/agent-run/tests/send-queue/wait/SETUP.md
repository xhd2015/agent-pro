# Scenario

**Feature**: send wait modes control blocking and timeout behavior

```
default -> poll until msg_N delivered
--no-wait -> return immediately after enqueue
--max-wait DURATION -> wall-clock timeout removes only caller message
```

## Preconditions

- Stub-tty scenario controls busy vs idle writable state.
- Queue file inspectable under `send-queue/<runner>/<session-id>.jsonl`.

## Steps

1. `Setup` sets `req.Operation = "wait"`.
2. Leaf selects busy, busy-then-idle, or permanently busy scenario via `req.Action`.
3. `Run` executes send with appropriate flags and records timing / queue state.
4. `Assert` checks delivery, timeout stderr, or fast return semantics.

```go
func Setup(t *testing.T, req *Request) error {
	req.Operation = "wait"
	req.EnableStubTTY = true
	return nil
}
```