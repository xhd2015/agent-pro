# Scenario

**Feature**: `tty status <session-id>` reads registry and reports status

```
agent-run tty status <session-id> -> registry file -> human-readable/JSON output
```

## Steps

1. Leaf `Setup` writes a mock registry entry for the session and sets `req.Args`.
2. `Run` executes `agent-run tty status ...`.
3. `Assert` checks exit code and output fields.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	writeMockRegistryEntry(t, req)
	return nil
}
```
