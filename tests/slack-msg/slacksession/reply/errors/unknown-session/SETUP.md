# Scenario

**Feature**: session reply with unknown session id

```
--session-id not-in-map MESSAGE -> session not found; exit 1
```

## Steps

1. Empty map (or map without id).
2. Args with unknown --session-id.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if err := seedSessionsJSON(t, req.HomeDir, []sessionMapEntry{}); err != nil {
		return err
	}
	req.Args = []string{
		"session", "reply",
		"--session-id", "slack-unknown-1",
		"--token", slackTestToken,
		"hello",
	}
	return nil
}
```
