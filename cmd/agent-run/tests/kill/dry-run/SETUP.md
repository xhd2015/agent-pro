# Scenario

**Feature**: `kill --dry-run` reports what would stop without terminating

```
live registry fixture
  -> agent-run kill --dry-run <session-id>
  -> stdout: dry-run: would stop <session-id>
  -> fixture PID still alive; registry still present
```

## Preconditions

- Live fixture helpers from kill root SETUP.
- Mode handle + AGENT_RUN_HOME isolation.

## Steps

1. Leaf starts live fixture and sets Args with `--dry-run`.
2. Run Handle.
3. Assert dry-run line, process still alive, registry intact.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	req.Mode = "handle"
	return nil
}
```
