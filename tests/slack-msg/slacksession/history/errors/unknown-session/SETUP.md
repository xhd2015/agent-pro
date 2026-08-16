# Scenario

**Feature**: session history unknown session

```
--session-id not-in-map -> session not found; exit 1
```

## Steps

1. Empty map.
2. Unknown --session-id.

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
		"session", "history",
		"--session-id", "slack-unknown-hist",
	}
	return nil
}
```
