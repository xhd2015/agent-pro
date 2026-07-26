# Scenario

**Feature**: same session id in both registries returns first provider

```
grok + codex both have session-1 -> registration order wins (grok first)
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "deterministic-order"
	return nil
}
```
