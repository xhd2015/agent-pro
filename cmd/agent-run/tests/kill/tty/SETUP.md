# Scenario

**Feature**: `tty kill` is an alias of top-level `kill`

```
agent-run tty kill <session-id> -> same stop contract as agent-run kill <session-id>
```

## Preconditions

- Live fixture helpers from kill root SETUP.
- Mode handle.

## Steps

1. Leaf starts fixture and sets Args to `tty kill <id>`.
2. Run Handle.
3. Assert same success path as top-level kill.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	req.Mode = "handle"
	return nil
}
```
