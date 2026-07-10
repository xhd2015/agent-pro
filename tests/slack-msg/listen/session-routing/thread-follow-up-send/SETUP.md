# Scenario

**Feature**: thread mode follow-up uses agent-run send

```
first message run --session -> second message in thread -> agent-run send <session-id>
```

## Steps

1. Inject two messages sharing thread_ts.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	threadTS := "1710000200.000100"
	req.WantAgentCalls = 2
	req.InjectEvents = []InjectedEvent{
		{
			Kind:    "app_mention",
			Channel: slackTestChannelID,
			Text:    "<@" + slackTestBotUserID + "> thread start",
			TS:      threadTS,
		},
		{
			Channel:  slackTestChannelID,
			Text:     "follow up in thread",
			TS:       "1710000200.000200",
			ThreadTS: threadTS,
		},
	}
	return nil
}
```
