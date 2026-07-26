# Scenario

**Feature**: provider RegistryDir follows id-registry convention

```
Get(grok-tty).RegistryDir -> grok-tty-registry
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "registry-dir-convention"
	req.RunnerID = "grok-tty"
	return nil
}
```
