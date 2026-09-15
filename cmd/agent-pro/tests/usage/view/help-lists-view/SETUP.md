# Scenario

**Feature**: the `usage` help lists the view sub-command

```
usage --help
Commands:
  collect   collect each provider's usage into ~/.agent-pro/usages (snapshots)
  list      list stored usage snapshots
  view      serve a read-only dashboard of the stored snapshots
exit 0
```

## Preconditions

- A user who never read this change learns the dashboard exists from this list, so
  the parent level has to name it.

## Steps

1. Ask the parent command for help.

```go
import (
"testing"

"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
req.Mode = ""
req.Command = []string{"usage"}
req.Args = []string{"--help"}
req.WithHomes = false
return nil
}
```
