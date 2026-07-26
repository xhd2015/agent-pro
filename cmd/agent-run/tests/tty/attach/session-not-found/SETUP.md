# Scenario

**Feature**: `tty attach` with unknown session id returns error

```
agent-run tty attach bogus-id -> registry file not found -> error
```

## Steps

1. `req.Args` = `["tty", "attach", "session-nonexistent"]` (no registry file).
2. Exit code 1, stderr mentions session not found.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"tty", "attach", "session-nonexistent"}
	return nil
}
```
