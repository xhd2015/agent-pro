# Scenario

**Feature**: whitespace-only session-id is rejected as missing

```
agent-run takeover "   " -> requires <session-id>; exit non-zero
```

## Steps

1. Args = `takeover` + whitespace-only positional.
2. Run Mode handle.
3. Assert non-zero exit and session-id / usage wording (not unknown command).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "handle"
	req.Args = []string{"takeover", "   "}
	return nil
}
```
