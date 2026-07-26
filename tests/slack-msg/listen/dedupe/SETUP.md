# Scenario

**Feature**: dedupe dual Slack events with same channel+ts

```
Socket Mode app_mention + message (same channel, ts) -> one agent launch
Socket Mode two messages different ts -> two agent launches
```

## Preconditions

- Default require-mention; channel @mention text includes bot user id.
- Daemon + mock agent-run.

## Steps

1. Start listen daemon with tokens.
2. Inject one or two events; assert launch count.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClearSlackEnv = true
	prependListenTokens(req)
	req.Daemon = true
	return nil
}
```
