# Scenario

**Feature**: live Slack send using explicit --config

```
slack-msg send --config repo/slack-config.json MESSAGE -> real botToken -> slack.com -> OK ts=... channel=...
```

## Preconditions

- Repo root `slack-config.json` with non-empty valid `botToken`.
- Network access to `slack.com`.
- Creates a real message in the configured default channel.

## Steps

1. Skip if repo config missing or empty token.
2. Run with `--config <abs-path>` and explicit message (no auto-discovery).

## Context

- Skipped unless `--label integration` or `--label slow` is passed.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	cfgPath := filepath.Join(req.RepoRoot, "slack-config.json")
	if _, err := os.Stat(cfgPath); err != nil {
		t.Skipf("skipping integration: %s not found", cfgPath)
	}
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Skipf("skipping integration: read config: %v", err)
	}
	if !strings.Contains(string(data), `"botToken"`) || strings.Contains(string(data), `"botToken": ""`) {
		t.Skipf("skipping integration: no live botToken in %s", cfgPath)
	}
	abs, err := filepath.Abs(cfgPath)
	if err != nil {
		return err
	}
	req.ConfigPath = abs
	req.SlackAPIURL = ""
	req.ClearSlackEnv = true
	return nil
}
```
