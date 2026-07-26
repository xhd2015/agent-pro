# Scenario

**Feature**: `tty status` human-readable output shows all registry fields

```
agent-run tty status session-1 -> stdout shows pid, port, tty type, session id, start time, tcp reachable
```

## Steps

1. `req.Args` = `["tty", "status", "session-1"]`.
2. Output should contain all expected human-readable fields.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"tty", "status", req.RegistrySessionID}
	return nil
}
```
