# Scenario

**Feature**: dry-run user dismisses/cancels the dialog

```
DEBUG_WITH_USER_DRY_RUN_DISMISSED=1 -> exit 1, no success JSON
```

## Steps

1. Enable dismissed simulation env var.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Env = append(req.Env, "DEBUG_WITH_USER_DRY_RUN_DISMISSED=1")
	return nil
}
```
