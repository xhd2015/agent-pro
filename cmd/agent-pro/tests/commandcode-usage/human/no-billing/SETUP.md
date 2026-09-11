# Scenario

**Feature**: missing billing endpoints render the empty state, not an error

```
# only whoami answers; every billing path 404s
FetchUsageWithOptions -> HasBillingData=false -> "No billing data found."
```

## Preconditions

- `whoami` answers, so the identity and Studio URL are known.
- `credits`, `subscriptions`, and the `usage/summary` are absent from the
  fixture map, so the fake API answers each with a 404 error envelope.

## Steps

1. Serve only the whoami body.

```go
import (
"testing"

"github.com/xhd2015/agent-pro/agent/commandcode"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
req.Bodies = map[string]string{
commandcode.PathWhoami: `{"success":true,"user":{"id":"u1","userName":"alice","name":"Alice","email":"alice@example.test"}}`,
}
return nil
}
```
