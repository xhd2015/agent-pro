# Scenario

**Feature**: `tty status` without session id returns error

```
agent-run tty status -> missing session id error + usage hint
```

## Steps

1. `req.Args` set to `["tty", "status"]` (no session id).
2. `Run` executes the command.
3. `Assert` checks exit code 1 and helpful error message.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"tty", "status"}
	return nil
}
```
