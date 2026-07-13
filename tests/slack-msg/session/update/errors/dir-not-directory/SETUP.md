# Scenario

**Feature**: session update --dir points at a file

```
--dir is a regular file -> dir is not a directory; exit 1
```

## Steps

1. Seed known session.
2. Create a file path; pass as --dir.

```go
import (
	"os"
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
	filePath := filepath.Join(req.WorkDir, "not-a-dir.txt")
	if err := os.WriteFile(filePath, []byte("x\n"), 0o644); err != nil {
		return err
	}
	req.Args = []string{
		"session", "update",
		"--session-id", sessionUpdateFixtureID,
		"--dir", filePath,
	}
	return nil
}
```
