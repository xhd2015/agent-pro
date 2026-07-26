# Scenario

**Feature**: `agent-run send` and `agent-run tty send` return same error for missing args

```
agent-run send -> same error as agent-run tty send (missing session-id and message)
```

## Steps

1. `req.Args` = `["send"]`.
2. Exit code 1, stderr mentions missing session id or message.
3. Error format matches what `tty send` produces with no args.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"send"}
	return nil
}
```