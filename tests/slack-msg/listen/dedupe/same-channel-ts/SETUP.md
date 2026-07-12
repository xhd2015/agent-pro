# Scenario

**Feature**: same channel+ts app_mention and message launch agent once

```
inject app_mention (C, ts, text) then message (C, ts, same text) -> 1 agent launch
```

## Steps

1. Inject dual events with identical channel, ts, and mention text.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	ts := "1710000500.000100"
	text := "<@" + slackTestBotUserID + "> hi from channel"
	req.WantAgentCalls = 1
	req.InjectEvents = []InjectedEvent{
		{
			Kind:    "app_mention",
			Channel: slackTestChannelID,
			User:    slackTestUserID,
			Text:    text,
			TS:      ts,
		},
		{
			Kind:    "message",
			Channel: slackTestChannelID,
			User:    slackTestUserID,
			Text:    text,
			TS:      ts,
		},
	}
	return nil
}
```
