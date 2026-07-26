# Scenario

**Feature**: dry-run Customize two-step flow yields free-text answer

```
step1 Customize -> step2 text field -> via=free_text, answer=typed report
```

## Steps

1. Stage step-1 button as `Customize`.
2. Stage step-2 text via `DEBUG_WITH_USER_DRY_RUN_TEXT`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Env = append(req.Env,
		"DEBUG_WITH_USER_DRY_RUN_BUTTON=Customize",
		"DEBUG_WITH_USER_DRY_RUN_TEXT=VS Code opened but wrong workspace",
	)
	return nil
}
```
