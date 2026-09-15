# Scenario

**Feature**: `usage list` reads the stored snapshots back, newest first

```
usage list [--last N] [--json]
  prints one row per stored record: ts | provider | usage | sessions | file
```

## Preconditions

- The store root comes from `AGENT_PRO_HOME`; there is no flag for it.
- Leaves plant snapshot files directly (`WriteSnapshot`), so the table's content
  is fixed and no collect has to run first.

## Steps

1. Set the command shape: `usage list` without home flags.
2. Leaves seed the store and pick `--last` or `--json`.

```go
import (
"testing"

"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
req.Command = []string{"usage", "list"}
req.WithHomes = false
return nil
}
```
