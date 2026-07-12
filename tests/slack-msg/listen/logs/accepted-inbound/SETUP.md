# Scenario

**Feature**: accepted inbound event is logged with user display

```
app_mention processed -> log kind + display name + channel + ts + quoted text
  -> agent open start log
```

## Steps

1. Inject one app_mention with fixed ts and distinctive text.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.WantAgentCalls = 1
	req.InjectEvents = []InjectedEvent{{
		Kind:    "app_mention",
		Channel: slackTestChannelID,
		User:    slackTestUserID,
		Text:    "<@" + slackTestBotUserID + "> log me please",
		TS:      "1710000600.000100",
	}}
	return nil
}
```
