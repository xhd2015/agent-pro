# Scenario

**Feature**: dry-run user picks a non-affirm preset button

```
DEBUG_WITH_USER_DRY_RUN_BUTTON=No — window did not open -> via=button, affirmed=false
```

## Steps

1. Stage button choice that does not match `--affirm`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Env = append(req.Env, "DEBUG_WITH_USER_DRY_RUN_BUTTON=No — window did not open")
	return nil
}
```
