# Scenario

**Feature**: `/api/series` returns in-window points, ascending per provider

```
usages/grok/...  used_percent 10 (50h ago, 1 session), 30 (2h ago, 3 sessions)
GET /api/series?metrics=used_percent,sessions_total&since=24h
  used_percent:  grok one point, the 2h one
  sessions_total: grok one point, drawn as a step
```

## Preconditions

- Two grok snapshots: one inside the 24h window and one well outside it, so a
  window that is ignored shows two points.
- `--since` is the run's default (7d); the query parameter is what narrows it, so
  the response also tells the dashboard which default produced it.

## Steps

1. Plant the two snapshots at 50h and 2h before now.
2. Probe `/api/series` for `used_percent` and `sessions_total` over 24h.

```go
import (
"testing"
"time"

"github.com/xhd2015/agent-pro/agent/usage"
"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
now := time.Now()
WriteViewSnapshot(t, req, usage.Grok, now.Add(-50*time.Hour),
map[string]float64{"used_percent": 10}, Display("10% used · monthly", "90% remaining"), 1, true, "")
WriteViewSnapshot(t, req, usage.Grok, now.Add(-2*time.Hour),
map[string]float64{"used_percent": 30}, Display("30% used · monthly", "70% remaining"), 3, true, "")
req.HTTPPath = "/api/series?metrics=used_percent,sessions_total&since=24h"
startViewServer(t, req)
return nil
}
```
