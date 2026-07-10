# Scenario

**Feature**: live Slack send using repo slack-config.json

```
slack-send (repo root cwd) -> real botToken -> slack.com -> OK ts=... channel=...
```

## Preconditions

- Repo root `slack-config.json` with non-empty valid `botToken`.
- Network access to `slack.com`.
- Creates a real message in the default channel (acceptable for debug script).

## Steps

1. Set `UseRepoConfig` true; run from `RepoRoot` with no args.
2. Assert success stdout shape with dynamic ts/config path placeholders.

## Context

- Skipped unless `--label integration` or `--label slow` is passed.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
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
	req.UseRepoConfig = true
	req.WriteGoMod = false
	req.ConfigFixture = ""
	req.ConfigInline = ""
	req.SlackAPIURL = ""
	req.Args = nil
	return nil
}
```