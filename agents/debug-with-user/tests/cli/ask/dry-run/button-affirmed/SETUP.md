# Scenario

**Feature**: dry-run user picks the affirm preset button

```
DEBUG_WITH_USER_DRY_RUN_BUTTON=Yes — window opened -> via=button, affirmed=true
```

## Steps

1. Stage button choice matching `--affirm`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Env = append(req.Env, "DEBUG_WITH_USER_DRY_RUN_BUTTON=Yes — window opened")
	return nil
}
```
