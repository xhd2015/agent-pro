# Scenario

**Feature**: `tty status` screen detection via ptywrap WebSocket scrollback analysis

```
# status reads scrollback from ptywrap WS, detects screen state
agent-run tty status session-1 -> WS connect -> scrollback -> banner/input-box/idle/error detection
```

## Preconditions

- A fake ptywrap HTTP+WebSocket server must be running on a reachable port.
- The registry entry points to the fake server.

## Steps

1. `Setup` starts a fake ptywrap WebSocket server, writes registry entry with the server's port.
2. Leaf `Setup` sets the scrollback content to simulate different screen states.
3. `Run` executes `agent-run tty status session-1`.
4. `Assert` checks that the screen status field matches the expected state.

```go
func Setup(t *testing.T, req *Request) error {
	req.StartFakePTYWrap = true
	startFakePTYWrapServer(t, req)
	waitForPortOpen(t, fmt.Sprintf("127.0.0.1:%d", req.FakePTYWrapPort), 5*time.Second)

	req.RegistryDir = "grok-tty-registry"
	req.RegistrySessionID = "session-1"
	req.RegistryEntryJSON = defaultRegistryEntryJSON(req.RegistrySessionID, req.FakePTYWrapPort)
	writeMockRegistryEntry(t, req)
	return nil
}
```
