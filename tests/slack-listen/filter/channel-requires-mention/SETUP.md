# Scenario

**Feature**: channel message without mention is ignored when requireMention default

```
channel message (no @bot) + default requireMention -> filtered
```

## Steps

1. Inject plain channel `message` without mention.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.WantAgentCalls = 0
	req.InjectEvents = []InjectedEvent{{
		Channel: slackTestChannelID,
		Text:    "hello channel without mention",
		TS:      "1710000003.000100",
	}}
	return nil
}
```