# Scenario

**Feature**: thread mode first open writes SYSTEM.md playbook

```
thread app_mention -> sessions/slack-C-ts/SYSTEM.md under ~/.agent-pro/slack-local-bot
  with slack-msg send / history recipes (no secrets)
```

## Steps

1. Isolate HOME under workdir.
2. Inject first thread message; wait for agent open.
3. Assert SYSTEM.md exists with recipes.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.WorkDir == "" {
		req.WorkDir = t.TempDir()
	}
	req.HomeDir = filepath.Join(req.WorkDir, "home")
	threadTS := "1710000700.000100"
	req.WantAgentCalls = 1
	req.InjectEvents = []InjectedEvent{{
		Kind:    "app_mention",
		Channel: slackTestChannelID,
		User:    slackTestUserID,
		Text:    "<@" + slackTestBotUserID + "> write system md",
		TS:      threadTS,
	}}
	return nil
}
```
