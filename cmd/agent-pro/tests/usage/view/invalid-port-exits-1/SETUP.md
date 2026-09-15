# Scenario

**Feature**: a port outside the TCP range is rejected before the server starts

```
usage view --port -1
stderr: agent-pro: invalid --port -1: must be between 0 and 65535
exit 1
```

## Preconditions

- This leaf runs the CLI in the foreground like the collect and list leaves do: a
  rejected flag never reaches the point where a server would block.
- `--port 0` is valid (any free port), so the check is a range check, not a
  truthiness one.

## Steps

1. Ask for a negative port.

```go
import (
"testing"

"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
req.Mode = ""
req.Command = []string{"usage", "view"}
req.Args = []string{"--port", "-1"}
req.WithHomes = false
return nil
}
```
