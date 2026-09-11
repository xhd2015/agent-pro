# Scenario

**Feature**: merged usage payload as JSON instead of the overlay

```
# the same UsageData that feeds ProjectUsageView marshals directly
UsageData -> FormatUsageJSON -> indented JSON
```

## Preconditions

- This branch only covers `req.Format = "json"`; width does not apply.

## Steps

1. Set `req.Format = "json"`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
req.Format = "json"
return nil
}
```
