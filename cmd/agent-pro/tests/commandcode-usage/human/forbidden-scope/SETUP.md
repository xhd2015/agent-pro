# Scenario

**Feature**: a scope the account may not read reports the API's permission
message instead of an expired session

```
# billing loads, but the summary answers 403 FORBIDDEN
whoami + credits + subscription(active) + summary(403)
  -> degraded overlay + the API's permission message
```

## Preconditions

- `whoami`, `credits`, and `subscriptions` answer normally.
- `usage/summary` answers 403 with a `FORBIDDEN` envelope, which is what the
  live API returns for an `orgId` the account cannot read.

## Steps

1. Serve the whole-account payload and replace only the summary.

```go
import (
"testing"
"time"

"github.com/xhd2015/agent-pro/agent/commandcode"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
req.Bodies = ActivePlanBodies(time.Now())
req.Bodies[commandcode.PathUsageSummary] = ErrorBody("FORBIDDEN", 403, "You do not have permission to view usage for this organization")
req.Statuses = map[string]int{commandcode.PathUsageSummary: 403}
return nil
}
```
