# Scenario

**Feature**: session update without --dir

```
session update --session-id ID (no --dir) -> nothing to update; exit 1
```

## Steps

1. Seed known session.
2. Args without --dir.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if err := seedSessionsJSON(t, req.HomeDir, []sessionMapEntry{{
		SessionID: sessionUpdateFixtureID,
		ChannelID: slackTestChannelID,
		Kind:      "channel",
		ReplyMode: "channel",
	}}); err != nil {
		return err
	}
	req.Args = []string{
		"session", "update",
		"--session-id", sessionUpdateFixtureID,
	}
	return nil
}
```
