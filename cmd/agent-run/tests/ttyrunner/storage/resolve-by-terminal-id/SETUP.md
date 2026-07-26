# Scenario

**Feature**: ResolveByTerminalID enriches from tty.json

```
registry session-1 + tty.json -> TTYSession with agent_session_id
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "resolve-by-terminal-id"
	return nil
}
```
