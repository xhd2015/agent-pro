# Scenario

**Feature**: thread interactive open does not PostMessage agent body

```
app_mention (default thread mode) -> RunInteractiveOpen -> 0 agent-body PostMessage
```

## Steps

1. Inject app_mention with fixed ts (default thread session mode).
2. Expect agent launch; no chat.postMessage of mock agent stdout.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.WantPosts = 0
	req.InjectEvents = []InjectedEvent{{
		Kind:    "app_mention",
		Channel: slackTestChannelID,
		Text:    "<@" + slackTestBotUserID + "> reply in thread please",
		TS:      "1710000400.000100",
	}}
	return nil
}
```
