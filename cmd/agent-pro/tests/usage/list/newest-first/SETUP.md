# Scenario

**Feature**: `usage list --last 2` shows the two newest snapshots, newest first

```
usages/grok/2026-09-15/09-00-00-snapshot.jsonl   10% used, 1 session
usages/codex/2026-09-15/10-00-00-snapshot.jsonl  20% used, 2 sessions
usages/grok/2026-09-15/11-00-00-snapshot.jsonl   30% used, 3 sessions
usage list --last 2
  2026-09-15T03:00:00Z  grok   30% used   3   .../11-00-00-snapshot.jsonl
  2026-09-15T02:00:00Z  codex  20% used   2   .../10-00-00-snapshot.jsonl
```

## Preconditions

- Three snapshots across two providers, planted with distinct seconds so both the
  ordering and the truncation are visible.
- The newest record is a grok one, so `--last` cannot pass by returning a whole
  provider's history.

## Steps

1. Plant the three snapshots.
2. List the two newest.

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
WriteSnapshot(t, req.Home, usage.Grok, "2026-09-15", "11-00-00",
time.Date(2026, 9, 15, 3, 0, 0, 0, time.UTC), 30, 3)
req.Args = []string{"--last", "2"}
return nil
}
```
