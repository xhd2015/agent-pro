# Scenario

**Bug**: `agent-run attach` and `agent-run tty attach` return same error for same missing session

```
agent-run attach session-xyz -> same error as tty attach session-xyz
```

## Steps

1. `req.Args` = `["attach", "session-xyz"]`.
2. Error format matches what `tty attach` produces for a different missing id.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"attach", "session-xyz"}
	return nil
}
```
