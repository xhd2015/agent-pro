# Scenario

**Feature**: empty URL leaves argv unchanged

```
AppendEventBusFlags(base, "", "") -> equal to base (copy, no flags added)
```

## Steps

1. Empty URL and token.
2. BaseArgs set by group/leaf.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.EventBusURL = ""
	req.EventBusToken = ""
	req.BaseArgs = []string{"run", "--auto-send-or-resume", "--session-id", "a1", "--", "hi"}
	return nil
}
```
