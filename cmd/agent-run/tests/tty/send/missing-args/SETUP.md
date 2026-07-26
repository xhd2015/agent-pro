# Scenario

**Feature**: `tty send` without required args returns error

```
agent-run tty send -> missing session-id and message
```

## Steps

1. `req.Args` = `["tty", "send"]`.
2. Exit code 1, stderr mentions missing args.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"tty", "send"}
	return nil
}
```
