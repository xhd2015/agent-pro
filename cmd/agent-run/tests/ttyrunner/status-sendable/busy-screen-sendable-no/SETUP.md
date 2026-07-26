# Scenario

**Feature**: codex working screen reports sendable no with reason

```
codex Working scrollback -> sendable: false + reason
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "busy-screen-sendable-no"
	req.RegistryDir = "codex-tty-registry"
	req.FakePTYWrapScrollback = "CODEX_TTY_BANNER\nCodex ›\n• Working on it (esc to interrupt)\n"
	return nil
}
```
