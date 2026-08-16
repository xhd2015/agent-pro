# Scenario

**Feature**: session reply validation errors

```
missing session id | unknown session | missing token/config -> stderr; exit 1
```

## Preconditions

- Clear host Slack / SLACK_MSG_* env.
- No network required.

## Steps

1. Clear env; isolate home when map needed.
2. Leaf sets invalid argv.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClearSlackEnv = true
	if err := isolateHome(t, req); err != nil {
		return err
	}
	return nil
}
```
