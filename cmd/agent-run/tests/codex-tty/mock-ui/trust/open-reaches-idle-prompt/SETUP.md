# Scenario

**Bug**: open leaves trust modal up; status reports `sendable: no (codex update available)`
because trust footer shares "Press enter to continue" with the update menu.

```
agent-run run --open --agent-runner codex-tty --agent-runner-binary llm-mock-run-codex
  -> trust dismissed
  -> sendable: yes (or idle › without trust)
```

## Steps

1. `Mode=mock-ui-open-idle` (empty prompt open).

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "mock-ui-open-idle"
	req.SessionID = "mock-ui-trust-idle"
	return nil
}
```
