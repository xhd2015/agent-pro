# Scenario

**Feature**: multiple observers all receive output

```
N observers -> all receive multiplexed PTY output
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "multiple-observers-all-receive"
	return nil
}
```
