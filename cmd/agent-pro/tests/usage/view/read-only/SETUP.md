# Scenario

**Feature**: serving a dashboard writes nothing to the store

```
usages/grok/<day>/<time>-snapshot.jsonl   planted line A
usages/codex/<day>/<time>-snapshot.jsonl  planted line B
usage view --port 0
GET /api/summary -> 200
store afterwards: the same two files, byte for byte
```

## Preconditions

- The store is planted by the leaf, so every byte on disk is known before the server
  starts and can be compared afterwards.
- `collect` owns the store; a viewing session must never look like a scrape, so a
  later `usage list` sees only collected snapshots.

## Steps

1. Plant two snapshots and remember their exact contents.
2. Serve, probe `/api/summary`, then compare the whole store tree.

```go
import (
"testing"
"time"

"github.com/xhd2015/agent-pro/agent/usage"
"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
now := time.Now()
WriteViewSnapshot(t, req, usage.Grok, now.Add(-2*time.Hour),
map[string]float64{"used_percent": 42, "remaining_percent": 58},
Display("42% used · weekly", "58% remaining"), 178, true, "")
WriteViewSnapshot(t, req, usage.Codex, now.Add(-time.Hour),
map[string]float64{"used_percent": 61}, Display("61% used", "39% remaining"), 92, true, "")
req.HTTPPath = "/api/summary"
startViewServer(t, req)
return nil
}
```
