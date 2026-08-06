# Scenario

**Feature**: help surfaces list and document `kill`

```
agent-run --help -> lists kill
agent-run kill --help -> usage, session-id, --dry-run
agent-run tty --help -> lists kill
```

## Preconditions

- L2 Mode handle (set by kill root Setup).
- No live registry fixture required.

## Steps

1. Leaf sets `req.Args` for the help invocation.
2. Run Handle in-process.
3. Assert exit 0 and required tokens / trailing newline.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	_ = d
	_ = req
	return nil
}
```
