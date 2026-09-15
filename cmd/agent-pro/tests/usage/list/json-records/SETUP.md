# Scenario

**Feature**: `usage list --json` emits the stored records verbatim

```
usage list --json
{"schema":"agent-pro/usage-snapshot/v1","ts":"2026-09-15T02:00:00Z","provider":"codex",...}
{"schema":"agent-pro/usage-snapshot/v1","ts":"2026-09-15T01:00:00Z","provider":"grok",...}
```

## Preconditions

- Two planted snapshots in two provider directories.
- The record shape is the one `usage collect --json` prints and the store writes,
  so a consumer can switch between collect and list output without changing its
  parser.

## Steps

1. Plant two snapshots.
2. List them as JSON.

```go
import (
"testing"
"time"

"github.com/xhd2015/agent-pro/agent/usage"
"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
WriteSnapshot(t, req.Home, usage.Grok, "2026-09-15", "09-00-00",
time.Date(2026, 9, 15, 1, 0, 0, 0, time.UTC), 10, 1)
WriteSnapshot(t, req.Home, usage.Codex, "2026-09-15", "10-00-00",
time.Date(2026, 9, 15, 2, 0, 0, 0, time.UTC), 20, 2)
req.Args = []string{"--json"}
return nil
}
```
