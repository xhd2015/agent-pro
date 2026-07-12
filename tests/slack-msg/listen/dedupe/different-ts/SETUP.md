# Scenario

**Feature**: different timestamps are not deduped

```
two app_mention events different ts -> 2 agent launches
```

## Steps

1. Inject two channel mentions with distinct ts values.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.WantAgentCalls = 2
	req.InjectEvents = []InjectedEvent{
		{
			Kind:    "app_mention",
			Channel: slackTestChannelID,
			User:    slackTestUserID,
			Text:    "<@" + slackTestBotUserID + "> first",
			TS:      "1710000510.000100",
		},
		{
			Kind:    "app_mention",
			Channel: slackTestChannelID,
			User:    slackTestUserID,
			Text:    "<@" + slackTestBotUserID + "> second",
			TS:      "1710000510.000200",
		},
	}
	return nil
}
```
