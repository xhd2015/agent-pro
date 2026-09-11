# Scenario

**Feature**: whoami's org and the subscription period scope every request

```
# a team namespace: whoami carries org_team, so billing calls are org-scoped
whoami(org=org_team) -> orgId=org_team -> credits/subscription/summary
subscription.currentPeriodStart -> since -> summary
```

## Preconditions

- whoami reports org `org_team` (login `acme`) with plan `teams-pro`.
- No `--org` or `--since`: both scopes must be derived from the responses.

## Steps

1. Serve a team payload where whoami carries the org.

```go
import (
"fmt"
"testing"
"time"

"github.com/xhd2015/agent-pro/agent/commandcode"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
now := time.Now()
req.Bodies = map[string]string{
commandcode.PathWhoami: `{"success":true,"user":{"id":"u1","userName":"alice","name":"Alice"},"org":{"id":"org_team","login":"acme","name":"Acme"}}`,
commandcode.PathCredits: `{"credits":{"belowThreshold":false,"creditThreshold":0,"monthlyCredits":8.5,"purchasedCredits":0,"freeCredits":0},"windowLimits":{"limited":false,"fiveHour":null,"weekly":null}}`,
commandcode.PathSubscriptions: fmt.Sprintf(
`{"success":true,"data":{"id":"sub_1","status":"active","planId":"teams-pro","quantity":1,"orgId":"org_team","currentPeriodStart":%q,"currentPeriodEnd":%q}}`,
now.Add(-10*24*time.Hour).Format(time.RFC3339), now.Add(20*24*time.Hour).Format(time.RFC3339)),
commandcode.PathUsageSummary: `{"totalCount":42,"totalCost":1.5,"averageCost":0.0357142857142857,"successRate":100,"completedCount":42,"failedCount":0,"periodBasis":"billing-period"}`,
}
return nil
}
```
