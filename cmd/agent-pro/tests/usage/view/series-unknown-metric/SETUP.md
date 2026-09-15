# Scenario

**Feature**: an unknown metric is an empty series, not an error

```
GET /api/series?metric=not_a_metric
  200 {"metrics":[{"name":"not_a_metric","series":[]}]}
```

## Preconditions

- The dashboard can ask for a metric a stored snapshot does not carry — a provider
  that stopped reporting a key, or a hand-typed URL.
- One grok snapshot exists, so the store is not the reason the series is empty.

## Steps

1. Plant one grok snapshot.
2. Probe `/api/series` for a metric no snapshot defines.

```go
import (
"testing"
"time"

"github.com/xhd2015/agent-pro/agent/usage"
"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
WriteViewSnapshot(t, req, usage.Grok, time.Now().Add(-time.Hour),
map[string]float64{"used_percent": 42}, Display("42% used", "58% remaining"), 178, true, "")
req.HTTPPath = "/api/series?metric=not_a_metric"
startViewServer(t, req)
return nil
}
```
