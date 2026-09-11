# Scenario

**Feature**: `--org` and `--since` override the derived scope

```
# both overrides set: whoami still supplies identity, not scope
orgId=org_pinned -> credits/subscription/summary
since=2026-01-01T00:00:00Z -> summary
```

## Preconditions

- whoami reports org `org_team`, so a derived scope would be `org_team`; the
  pinned `org_pinned` must win.
- The subscription reports its own period start, which the pinned `since` must
  win over.

## Steps

1. Serve the team payload under the pinned overrides.

```go
import (
"fmt"
"testing"
"time"

"github.com/xhd2015/agent-pro/agent/commandcode"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
now := time.Now()
req.Org = "org_pinned"
req.Since = "2026-01-01T00:00:00Z"
req.Bodies = map[string]string{
commandcode.PathWhoami: `{"success":true,"user":{"id":"u1","userName":"alice","name":"Alice"},"org":{"id":"org_team","login":"acme","name":"Acme"}}`,
commandcode.PathCredits: `{"credits":{"belowThreshold":false,"creditThreshold":0,"monthlyCredits":3.25,"purchasedCredits":0,"freeCredits":0},"windowLimits":{"limited":false,"fiveHour":null,"weekly":null}}`,
commandcode.PathSubscriptions: fmt.Sprintf(
`{"success":true,"data":{"id":"sub_1","status":"active","planId":"teams-pro","quantity":1,"orgId":"org_team","currentPeriodStart":%q,"currentPeriodEnd":%q}}`,
now.Add(-10*24*time.Hour).Format(time.RFC3339), now.Add(20*24*time.Hour).Format(time.RFC3339)),
commandcode.PathUsageSummary: `{"totalCount":7,"totalCost":0.5,"averageCost":0.07142857142857142,"successRate":100,"completedCount":7,"failedCount":0,"periodBasis":"custom"}`,
}
return nil
}
```
