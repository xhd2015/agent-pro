# Scenario

**Feature**: unreachable terminal reports sendable false

```
registry with closed port -> sendable: false, unreachable
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "unreachable-sendable-no"
	req.StartFakePTYWrap = false
	req.RegistryDir = "grok-tty-registry"
	writeRegistryEntry(t, req.Home, req.RegistryDir, "session-1", defaultRegistryEntryJSON("session-1", 59997))
	return nil
}
```
