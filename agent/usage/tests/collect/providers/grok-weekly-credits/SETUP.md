# Scenario

**Feature**: grok reports its weekly credit period when the monthly billing payload is absent

```
delete fixture grok-billing.json
Collect -> grok /v1/billing 404 fixture -> ?format=credits -> 2% of the week used
```

## Preconditions

- Inherits the root fixtures minus `grok-billing.json`, the file backing
  `/v1/billing`.
- The credits fixture describes a `USAGE_PERIOD_TYPE_WEEKLY` period with
  `creditUsagePercent: 2`.

## Steps

1. Remove `grok-billing.json` so only the credits payload can answer.

```go
import (
"testing"

"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
delete(req.Fixtures, "grok-billing.json")
return nil
}
```
