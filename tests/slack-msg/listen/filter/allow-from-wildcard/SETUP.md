# Scenario

**Feature**: default allowFrom * accepts any user

```
default allowFrom + app_mention from any user -> agent invoked
```

## Steps

1. Inject app_mention from slackTestOtherUserID (default allowFrom *).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.WantAgentCalls = 1
	req.InjectEvents = []InjectedEvent{{
		Kind:    "app_mention",
		Channel: slackTestChannelID,
		User:    slackTestOtherUserID,
		Text:    "<@" + slackTestBotUserID + "> wildcard allow",
		TS:      "1710000006.000100",
	}}
	return nil
}
```
