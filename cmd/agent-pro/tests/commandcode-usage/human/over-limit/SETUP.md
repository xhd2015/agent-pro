# Scenario

**Feature**: a saturated account renders near-100% usage without a full bar

```
# 0.06 credits left of the 10-credit pool is 99.4% used
monthlyCredits 0.06 + both windows at cap -> ProjectUsageView -> FormatUsage
```

## Preconditions

- 0.06 monthly credits remain of the 10-credit `individual-go` allowance, so
  99.4% is used: rounding alone would fill every block.
- The 5-hour window is at 3/3 and the weekly window at 6/6, both flagged
  `exceeded`, and the weekly reset is 2h 30m out.

## Steps

1. Serve a saturated payload inline.

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
commandcode.PathWhoami: `{"success":true,"user":{"id":"u1","userName":"alice","name":"Alice","email":"alice@example.test"}}`,
commandcode.PathCredits: fmt.Sprintf(
`{"credits":{"belowThreshold":false,"creditThreshold":0,"monthlyCredits":0.06,"purchasedCredits":0,"freeCredits":0},"windowLimits":{"limited":true,"fiveHour":{"used":3,"cap":3,"exceeded":true,"resetAt":0},"weekly":{"used":6,"cap":6,"exceeded":true,"resetAt":%d}}}`,
now.Add(2*time.Hour+30*time.Minute).UnixMilli()),
commandcode.PathSubscriptions: fmt.Sprintf(
`{"success":true,"data":{"id":"sub_1","status":"active","planId":"individual-go","quantity":1,"currentPeriodStart":%q,"currentPeriodEnd":%q}}`,
now.Add(-28*24*time.Hour).Format(time.RFC3339), now.Add(5*24*time.Hour).Format(time.RFC3339)),
commandcode.PathUsageSummary: `{"totalCount":77,"totalCost":9.94,"averageCost":0.1290909090909091,"successRate":100,"completedCount":77,"failedCount":0,"periodBasis":"billing-period"}`,
}
return nil
}
```
