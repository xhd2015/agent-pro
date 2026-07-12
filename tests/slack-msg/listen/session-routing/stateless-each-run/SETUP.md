# Scenario

**Feature**: stateless mode uses agentrunbridge Run (capture) for every message

```
--session-mode stateless + two messages -> two run launches (no send, no --open)
```

## Steps

1. Pass `--session-mode stateless`.
2. Inject two app_mentions.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = append(req.Args, "--session-mode", "stateless")
	req.WantAgentCalls = 2
	req.InjectEvents = []InjectedEvent{
		{
			Kind:    "app_mention",
			Channel: slackTestChannelID,
			Text:    "<@" + slackTestBotUserID + "> one",
			TS:      "1710000300.000100",
		},
		{
			Kind:    "app_mention",
			Channel: slackTestChannelID,
			Text:    "<@" + slackTestBotUserID + "> two",
			TS:      "1710000300.000200",
		},
	}
	return nil
}
```
