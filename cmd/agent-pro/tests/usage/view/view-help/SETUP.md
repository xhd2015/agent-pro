# Scenario

**Feature**: `usage view --help` documents the flags and the exit contract

```
usage view --help
Usage: agent-pro usage view [options]
Options:
  --port <n>  --open  --no-open  --since <dur>  --static-dir <d>  --json
Exit status:
  0 on a clean shutdown; 1 on a bad flag, an unreadable store root, or no free port
```

## Preconditions

- The flags a user needs to find are the ones that decide where the dashboard runs
  and what it shows: the port, the browser, the range, the static override.

## Steps

1. Ask the sub-command for help.

```go
import (
"testing"

"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
req.Mode = ""
req.Command = []string{"usage", "view"}
req.Args = []string{"--help"}
req.WithHomes = false
return nil
}
```
