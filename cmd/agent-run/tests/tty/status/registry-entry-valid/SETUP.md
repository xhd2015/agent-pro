# Scenario

**Feature**: `tty status <session-id>` with valid grok-tty registry entry

```
agent-run tty status session-1 -> grok-tty-registry/session-1.json -> status fields
```

## Steps

1. Leaf `Setup` writes a grok-tty mock registry entry and sets `req.Args`.
2. `Run` executes `agent-run tty status session-1`.
3. `Assert` checks exit code and output fields.

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
