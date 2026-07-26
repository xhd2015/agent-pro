# Scenario

**Feature**: ResolveByTerminalID / LookupSession across provider registries

```
LookupSession(home, terminal-id) -> search providers in order -> skip stale -> return entry
```

## Preconditions

- Registry JSON files under `<runner-id>-registry/`.
- TCP reachability determines stale vs live entries.

## Steps

1. Leaf sets `req.Operation = "lookup"` and writes registry fixtures.
2. `Run` calls `ttyrunner.LookupSession`.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Operation = "lookup"
	req.RegistrySessionID = "session-1"
	return nil
}
```
