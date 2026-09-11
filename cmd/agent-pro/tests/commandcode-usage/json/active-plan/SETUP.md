# Scenario

**Feature**: a full payload marshals with the cmd CLI's key shape

```
# every endpoint answered, so every top-level key is an object
whoami + credits + subscription + summary -> FormatUsageJSON
```

## Preconditions

- Billing data loads and the subscription is `active` with plan `individual-go`.
- The credits payload keeps sub-cent precision: 2.408591571 monthly credits.

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
