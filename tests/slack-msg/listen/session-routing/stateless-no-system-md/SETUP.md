# Scenario

**Feature**: stateless mode does not write SYSTEM.md

```
--session-mode stateless + app_mention -> agent Run capture
  -> no ~/.agent-pro/slack-local-bot/sessions/.../SYSTEM.md
```

## Steps

1. Isolate HOME; pass `--session-mode stateless`.
2. Inject one mention; assert agent ran and SYSTEM.md absent.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.WorkDir == "" {
		req.WorkDir = t.TempDir()
	}
	req.HomeDir = filepath.Join(req.WorkDir, "home")
	req.Args = append(req.Args, "--session-mode", "stateless")
	req.WantAgentCalls = 1
	req.InjectEvents = []InjectedEvent{{
		Kind:    "app_mention",
		Channel: slackTestChannelID,
		User:    slackTestUserID,
		Text:    "<@" + slackTestBotUserID + "> stateless no system",
		TS:      "1710000720.000100",
	}}
	return nil
}
```
