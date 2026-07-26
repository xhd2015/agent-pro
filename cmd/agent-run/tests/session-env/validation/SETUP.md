# Scenario

**Feature**: invalid `-e`/`--env` values are rejected before spawn

```
-e FOO   (no =)  -> exit ≠ 0; clear error
-e =bar  (empty key) -> exit ≠ 0; clear error
```

## Steps

1. Leaves pass invalid env flags on TTY `run` with a prompt.
2. Assert non-zero exit and a clear error message.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Validation leaves override Args fully; baseline is TTY run prefix.
	if len(req.Args) == 0 {
		req.Args = []string{"run", "--agent-runner", "grok-tty"}
	}
	return nil
}
```
