# Scenario

**Feature**: thread RunInteractiveOpen passes SLACK_MSG_SESSION_ID + SLACK_MSG_CONFIG via -e

```
listen --config PATH + app_mention
  -> agent-run argv includes:
     -e SLACK_MSG_SESSION_ID=slack-channel-{channelID}
     -e SLACK_MSG_CONFIG=<abs config path>
```

## Steps

1. Isolate HOME; materialize config for abs path.
2. Inject thread message; wait for agent open.
3. Assert INVOCATION argv contains both env flags.

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
	if err := withConfigArg(t, d, req, "valid-config.json", false); err != nil {
		return err
	}
	threadTS := "1710000740.000100"
	req.WantAgentCalls = 1
	req.InjectEvents = []InjectedEvent{{
		Kind:    "app_mention",
		Channel: slackTestChannelID,
		User:    slackTestUserID,
		Text:    "<@" + slackTestBotUserID + "> env please",
		TS:      threadTS,
	}}
	return nil
}
```
