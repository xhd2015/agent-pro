# Scenario

**Feature**: `tty attach <session-id>` connects to live TTY session via registry

```
agent-run tty attach <session-id> -> registry lookup -> ptyclient.Attach -> interactive WS
```

## Preconditions

- A fake ptywrap WebSocket server is running (for valid-connection tests).
- Registry entry points to the reachable server.

## Steps

1. Leaf `Setup` writes mock registry entry and optionally starts a fake ptywrap server.
2. `Run` executes CLI or probes the attach WS connection.
3. `Assert` checks exit code or probe result.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.RegistryDir = "grok-tty-registry"
	req.RegistrySessionID = "session-1"
	writeMockRegistryEntry(t, req)
	return nil
}
```
