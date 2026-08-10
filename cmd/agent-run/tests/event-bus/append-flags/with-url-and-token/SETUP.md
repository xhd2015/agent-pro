# Scenario

**Feature**: URL set appends both event-bus flags (token when non-empty)

```
AppendEventBusFlags(base, url, token)
  -> … base … --event-bus-url <url> --event-bus-token <token>
```

## Steps

1. Set URL and token.
2. BaseArgs without event-bus flags.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.EventBusURL = "http://127.0.0.1:23891"
	req.EventBusToken = "tok-a2"
	req.BaseArgs = []string{"run", "--auto-send-or-resume", "--session-id", "a2"}
	return nil
}
```
