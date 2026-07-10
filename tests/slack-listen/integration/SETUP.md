# Scenario

**Feature**: live Socket Mode bridge with real tokens and agent-run

```
slack-listen listen --config repo/slack-config.json -> live Socket Mode -> agent-run -> Slack reply
```

## Preconditions

- Repo root `slack-config.json` with non-empty valid `botToken` and `appToken`.
- Network access to `slack.com`.
- Manual or scripted inbound Slack message during probe window (or implementer test hook).

## Steps

1. Skip if repo config missing or tokens empty.
2. Run daemon with explicit `--config` and short observe timeout.
3. Assert process connects and logs config path (full E2E may require human message).

## Context

- Skipped unless `--label integration` or `--label slow` is passed.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	cfgPath := filepath.Join(req.RepoRoot, "slack-config.json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Skipf("skipping integration: %s not found", cfgPath)
	}
	content := string(data)
	if !strings.Contains(content, `"botToken"`) || strings.Contains(content, `"botToken": ""`) {
		t.Skipf("skipping integration: no live botToken in %s", cfgPath)
	}
	if !strings.Contains(content, `"appToken"`) || strings.Contains(content, `"appToken": ""`) {
		t.Skipf("skipping integration: no live appToken in %s", cfgPath)
	}
	abs, err := filepath.Abs(cfgPath)
	if err != nil {
		return err
	}
	req.ConfigPath = abs
	req.SlackAPIURL = ""
	req.ClearSlackEnv = true
	req.BotToken = ""
	req.AppToken = ""
	req.Daemon = true
	req.WantAgentCalls = 0
	req.ObserveTimeout = 15 * time.Second
	return nil
}
```