# Scenario

**Feature**: session update success (set dir / json)

```
seed map entry with non-empty fields
  -> session update --session-id ID --dir PATH [--json]
  -> abs dir stored; other fields preserved; updated_at bumped
```

## Preconditions

- Isolated HomeDir; clear Slack env.
- Workspace directory created under WorkDir before invoke.

## Steps

1. Isolate home; seed map entry with distinctive fields (no dir yet).
2. Leaf creates dir path and sets update args.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func sessionUpdateFixtureEntry() sessionMapEntry {
	return sessionMapEntry{
		SessionID:          sessionUpdateFixtureID,
		ChannelID:          slackTestChannelID,
		ThreadTS:           "1710000700.000100",
		ConfigPath:         "/tmp/slack-update-cfg.json",
		Dir:                "",
		Kind:               "channel",
		ReplyMode:          "channel",
		LastMessagePreview: "before update",
		CreatedAt:          "2026-07-09T10:00:00Z",
		UpdatedAt:          "2026-07-10T10:00:00Z",
	}
}

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	if err := isolateHome(t, req); err != nil {
		return err
	}
	if err := seedSessionsJSON(t, req.HomeDir, []sessionMapEntry{sessionUpdateFixtureEntry()}); err != nil {
		return err
	}
	return nil
}

func ensureUpdateWorkspaceDir(t *testing.T, req *Request) (string, error) {
	t.Helper()
	dir := filepath.Join(req.WorkDir, "agent-workspace")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	return abs, nil
}
```
