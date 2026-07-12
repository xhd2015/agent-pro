# Scenario

**Feature**: thread mode appends inbound message to messages.jsonl

```
app_mention text "log me please" -> sessions/slack-channel-{channelID}/messages.jsonl
  {"message_id"|"ts", user, text, direction:"in"}
```

## Steps

1. Isolate HOME.
2. Inject one accepted inbound event.
3. Assert messages.jsonl has inbound line with text.

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
	threadTS := "1710000730.000100"
	req.WantAgentCalls = 1
	req.InjectEvents = []InjectedEvent{{
		Kind:    "app_mention",
		Channel: slackTestChannelID,
		User:    slackTestUserID,
		Text:    "<@" + slackTestBotUserID + "> log me please",
		TS:      threadTS,
	}}
	return nil
}
```
