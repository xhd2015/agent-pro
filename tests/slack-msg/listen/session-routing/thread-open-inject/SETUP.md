# Scenario

**Feature**: thread open inject prompt strips bot mention and includes SYSTEM.md path

```
app_mention text "<@BOT> clean me" -> agent INVOCATION prompt:
  Slack listen session open / session-id / channel / thread_ts / from: / Instructions: / User message:
  without raw <@BOTID>; with stripped "clean me"
```

## Steps

1. Isolate HOME for SYSTEM.md path.
2. Inject app_mention with bot mention prefix.
3. Assert INVOCATION line content.

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
	threadTS := "1710000710.000100"
	req.WantAgentCalls = 1
	req.InjectEvents = []InjectedEvent{{
		Kind:    "app_mention",
		Channel: slackTestChannelID,
		User:    slackTestUserID,
		Text:    "<@" + slackTestBotUserID + "> clean me",
		TS:      threadTS,
	}}
	return nil
}
```
