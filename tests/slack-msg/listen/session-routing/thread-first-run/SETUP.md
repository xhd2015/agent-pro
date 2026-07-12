# Scenario

**Feature**: thread mode first message uses agentrunbridge RunInteractiveOpen

```
first thread message -> agent-run run --session-id=slack-C-ts
  --auto-send-or-resume --new-terminal --open -- <prompt>
```

## Steps

1. Inject app_mention establishing thread root ts.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	threadTS := "1710000100.000100"
	req.WantAgentCalls = 1
	req.InjectEvents = []InjectedEvent{{
		Kind:    "app_mention",
		Channel: slackTestChannelID,
		Text:    "<@" + slackTestBotUserID + "> start thread",
		TS:      threadTS,
	}}
	return nil
}
```
