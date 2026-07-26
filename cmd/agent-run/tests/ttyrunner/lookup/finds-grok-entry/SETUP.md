# Scenario

**Feature**: LookupSession finds grok-tty registry entry

```
grok-tty-registry/session-1.json reachable -> entry + grok-tty
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "finds-grok-entry"
	return nil
}
```
