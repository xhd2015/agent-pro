# Scenario

**Feature**: `agent-run send` alias shares queue semantics with `agent-run tty send`

```
agent-run send <id> "msg" -> same agentsend queue as agent-run tty send <id> "msg"
```

## Preconditions

- Idle stub-tty background session.

## Steps

1. `Setup` sets `req.Operation = "alias"`.
2. `Run` enqueues via `send` then `tty send` on same session.
3. `Assert` both print monotonic msg ids (`msg_1`, `msg_2`) with exit 0.

```go
func Setup(t *testing.T, req *Request) error {
	req.Operation = "alias"
	req.EnableStubTTY = true
	return nil
}
```