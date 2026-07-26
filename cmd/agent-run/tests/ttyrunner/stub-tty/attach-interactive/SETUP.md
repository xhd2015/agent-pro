# Scenario

**Feature**: interactive attach to live stub-tty session

```
attach_mode=interactive -> writer can send input
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "attach-interactive"
	return nil
}
```
