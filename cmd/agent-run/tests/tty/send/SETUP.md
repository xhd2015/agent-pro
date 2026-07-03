# Scenario

**Feature**: `tty send <session-id> "msg"` injects prompt into live terminal and captures response

```
# resolve terminal via registry -> connect WS -> inject prompt -> capture response -> append events
agent-run tty send <session-id> "msg" -> resolveTerminal -> WS inject -> capture -> store events
```

## Preconditions

- A fake ptywrap WebSocket server is running for send-with-server tests.
- Registry entry points to the reachable server.

## Steps

1. `Setup` writes mock registry entry and optionally starts a fake ptywrap server.
2. Leaf `Setup` sets `req.Args` for the send command.
3. `Run` executes `agent-run tty send <session-id> "msg"`.
4. `Assert` checks exit code and response capture.

```go
func Setup(t *testing.T, req *Request) error {
	req.RegistryDir = "grok-tty-registry"
	req.RegistrySessionID = "session-1"
	writeMockRegistryEntry(t, req)
	return nil
}
```
