# Scenario

**Feature**: ttyrunner.IDs lists builtin providers

```
ttyrunner.IDs() -> [grok-tty, codex-tty, stub-tty?]
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "lists-builtin-providers"
	return nil
}
```
