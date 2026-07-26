# Scenario

**Feature**: unified session storage — dual-write tty.json and resolver API

```
TTY run -> registry JSON + sessions/<runner>/<agent-id>/tty.json
ResolveByAgentSession / ResolveByTerminalID -> enriched TTYSession
```

## Preconditions

- Resolver reads registry dirs in provider registration order.
- `tty.json` cross-references `terminal_session_id`, `listen_addr`, `alive`.

## Steps

1. Leaf sets `req.Operation = "storage"` and `req.Action`.
2. Fixtures write registry, meta, and tty.json as needed.
3. `Run` exercises dual-write or resolver methods.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Operation = "storage"
	req.EnableStubTTY = true
	return nil
}
```
