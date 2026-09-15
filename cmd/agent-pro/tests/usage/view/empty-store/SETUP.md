# Scenario

**Feature**: an empty store still serves a dashboard

```
AGENT_PRO_HOME=<tmp>            # no usages tree at all
usage view --port 0
stdout: serving http://127.0.0.1:<port>
        store   <tmp>/usages  (read-only, 0 providers, 0 snapshots)
GET /api/summary -> {"providers":[],"recent":[],...}
```

## Preconditions

- A first-time user starts the dashboard before any collect ran: the store tree does
  not exist.
- The page itself tells that user to run `agent-pro usage collect` (checked by the
  package test, which serves the same embedded page).

## Steps

1. Plant nothing.
2. Probe `/api/summary`.

```go
import (
"testing"

"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
req.HTTPPath = "/api/summary"
startViewServer(t, req)
return nil
}
```
