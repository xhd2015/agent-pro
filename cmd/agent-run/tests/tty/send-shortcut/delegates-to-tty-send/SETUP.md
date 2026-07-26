# Scenario

**Feature**: `agent-run send` and `agent-run tty send` produce identical error for missing session

```
agent-run send bogus-id "hello" vs agent-run tty send bogus-id "hello" -> same error message
```

## Steps

1. `req.Args` = `["send", "session-nonexistent", "hello"]`.
2. Exit code 1, error mentions not found.
3. Stderr output should match the tty send equivalent.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"send", "session-nonexistent", "hello"}
	return nil
}
```