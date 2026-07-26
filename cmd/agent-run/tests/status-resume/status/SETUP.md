# Scenario

**Feature**: `agent-run status` bare home and session multi-layer probe

```
agent-run status -> home: <path>
agent-run status <session-id|runner/session> [--json] -> multi-layer probe
agent-run status --grok-session-id ID [--json] -> meta-only resolve then probe
```

## Steps

1. Leaf seeds meta/registry as needed and sets `req.Args`.
2. `Run` executes status; assert exit code and layer fields.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Default to bare status; leaves override Args and seeds.
	if len(req.Args) == 0 {
		req.Args = []string{"status"}
	}
	return nil
}
```
