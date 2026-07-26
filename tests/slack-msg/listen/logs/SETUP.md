# Scenario

**Feature**: operator logs for inbound events and agent open outcomes

```
accepted inbound -> stderr summary (kind, user display, channel, ts, text)
agent open start / failure -> stderr lines (not silent)
```

## Preconditions

- Daemon + mock agent; slacktest users.info provides display name `spengler`.

## Steps

1. Start listen; inject processable event.
2. Assert stderr contains operator-facing log lines.

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
