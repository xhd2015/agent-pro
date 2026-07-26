# Scenario

**Feature**: second interactive attach is read-only

```
first interactive -> writer; second interactive -> observer (input dropped)
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "first-attach-writes-second-readonly"
	return nil
}
```
