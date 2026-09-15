# Scenario

**Feature**: `/api/summary` reports the newest snapshot per provider

```
usages/grok/...          used_percent 30 (3h ago), 42 (1h ago)
usages/codex/...         used_percent 61 (2h ago)
usages/commandcode/...   usage_percent 29 (30m ago)
GET /api/summary
  providers: codex 61% · commandcode 29% · grok 42% (2 snapshots)
  recent:    newest first, across providers
  metrics:   the menu the dashboard builds its selector from
```

## Preconditions

- Two grok snapshots and one each for codex and commandcode: grok's card must show
  the newest value, not the first file read, and its count must be 2.
- Command Code reports `usage_percent`, not `used_percent`, so the normalized
  `primary_percent` metric has to map provider keys.

## Steps

1. Plant the four snapshots, newest instants weeks apart from the previous day.
2. Probe `/api/summary`.

```go
import (
"testing"
"time"

"github.com/xhd2015/agent-pro/agent/usage"
"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
now := time.Now()
WriteViewSnapshot(t, req, usage.Grok, now.Add(-3*time.Hour),
map[string]float64{"used_percent": 30, "remaining_percent": 70},
Display("30% used · weekly", "70% remaining"), 100, true, "")
WriteViewSnapshot(t, req, usage.Grok, now.Add(-time.Hour),
map[string]float64{"used_percent": 42, "remaining_percent": 58},
Display("42% used · weekly", "58% remaining"), 178, true, "")
WriteViewSnapshot(t, req, usage.Codex, now.Add(-2*time.Hour),
map[string]float64{"used_percent": 61, "remaining_percent": 39},
Display("61% used", "39% remaining"), 92, true, "")
WriteViewSnapshot(t, req, usage.CommandCode, now.Add(-30*time.Minute),
map[string]float64{"usage_percent": 29, "cost_usd": 12.25},
Display("$12.25 spent · Go", "66 credits remaining"), 41, true, "")
req.HTTPPath = "/api/summary"
startViewServer(t, req)
return nil
}
```
