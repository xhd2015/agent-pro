# Scenario

**Feature**: session update --dir path does not exist

```
--dir missing path -> dir does not exist; exit 1
```

## Steps

1. Seed known session.
2. --dir points to non-existent path.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if err := seedSessionsJSON(t, req.HomeDir, []sessionMapEntry{{
		SessionID: sessionUpdateFixtureID,
		ChannelID: slackTestChannelID,
		Kind:      "channel",
		ReplyMode: "channel",
	}}); err != nil {
		return err
	}
	missing := filepath.Join(req.WorkDir, "no-such-workspace-dir")
	req.Args = []string{
		"session", "update",
		"--session-id", sessionUpdateFixtureID,
		"--dir", missing,
	}
	return nil
}
```
