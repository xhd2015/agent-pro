# Scenario

**Feature**: dry-run returns candidate without focusing

```
DryRun=true + one candidate -> FocusSession ok; FocusITerm never called
```

## Steps

1. Inherited single-match fixtures; set DryRun.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.DryRun = true
	return nil
}
```
