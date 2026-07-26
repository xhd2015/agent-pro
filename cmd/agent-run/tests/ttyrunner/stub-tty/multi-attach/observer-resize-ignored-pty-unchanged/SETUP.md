# Scenario

**Feature**: observer resize is dropped

```
observer resize JSON -> ignored (no PTY dimension change)
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "observer-resize-ignored-pty-unchanged"
	return nil
}
```
