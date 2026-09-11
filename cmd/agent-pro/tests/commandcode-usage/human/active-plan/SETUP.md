# Scenario

**Feature**: a full individual-go payload renders the CLI overlay line for line

```
# an active subscription bounds the pool at the 10-credit plan allowance
whoami + credits($2.41 left, 0/3 and 0.48/6 windows) + subscription(active)
  + summary(1040 requests) -> ProjectUsageView -> FormatUsage
```

## Preconditions

- Billing data loads and the subscription is `active` with plan `individual-go`.
- 2.408591571 monthly credits remain, so 75.914% of the 10-credit pool is used.
- The weekly window resets 7h 4m out; the billing period ends 5 days out.

## Steps

1. Serve `ActivePlanBodies`.

```go
import (
"testing"
"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
req.Bodies = ActivePlanBodies(time.Now())
return nil
}
```
