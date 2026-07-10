# Scenario

**Feature**: reply uses thread_ts from inbound message

```
app_mention with ts -> PostMessage thread_ts matches message ts
```

## Steps

1. Inject app_mention with fixed ts.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.InjectEvents = []InjectedEvent{{
		Kind:    "app_mention",
		Channel: slackTestChannelID,
		Text:    "<@" + slackTestBotUserID + "> reply in thread please",
		TS:      "1710000400.000100",
	}}
	return nil
}
```
