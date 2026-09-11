# Scenario

**Feature**: request scoping — which org and period the fetch derives

```
# the composite fetch threads org and period through every endpoint
whoami.org.id -> orgId -> credits/subscription/summary
subscription.currentPeriodStart -> since -> summary
```

## Preconditions

- This branch covers the `--org` / `--since` overrides: with neither set the
  scope is derived, and either one wins over the derived value.
- `req.Queries` records the raw query string per path for the leaves to assert.

## Steps

1. Pin the terminal width so the leaves may also check rendered output.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
req.Format = "human"
req.Width = 80
return nil
}
```
