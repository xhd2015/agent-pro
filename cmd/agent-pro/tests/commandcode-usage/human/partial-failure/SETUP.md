# Scenario

**Feature**: endpoint failures degrade the overlay instead of aborting it

```
# credits 401 and summary 502, but the subscription still loads
FetchUsageWithOptions -> HasBillingData=true, Credits=nil -> degraded overlay
```

## Preconditions

- `whoami` and `subscriptions` answer with the active `individual-go` payload
  (period start 2026-08-14, period end 2026-09-14).
- `credits` answers 401 with an `UNAUTHORIZED` envelope and `usage/summary`
  answers 502 with a `BAD_GATEWAY` envelope.

## Steps

1. Serve the mixed payload with per-path failure statuses.

```go
import (
"testing"

"github.com/xhd2015/agent-pro/agent/commandcode"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
req.Bodies = map[string]string{
commandcode.PathWhoami: `{"success":true,"user":{"id":"u1","userName":"alice","name":"Alice","email":"alice@example.test"}}`,
commandcode.PathCredits:       ErrorBody("UNAUTHORIZED", 401, "invalid api key"),
commandcode.PathSubscriptions: `{"success":true,"data":{"id":"sub_1","status":"active","planId":"individual-go","quantity":1,"currentPeriodStart":"2026-08-14T00:05:12.000Z","currentPeriodEnd":"2026-09-14T00:05:12.000Z"}}`,
commandcode.PathUsageSummary:  ErrorBody("BAD_GATEWAY", 502, "upstream unavailable"),
}
req.Statuses = map[string]int{
commandcode.PathCredits:      401,
commandcode.PathUsageSummary: 502,
}
return nil
}
```
