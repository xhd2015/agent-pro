# Scenario

**Feature**: --reply-prefix prepended to Slack reply

```
--reply-prefix [bot] + agent output -> PostMessage text starts with prefix
```

## Steps

1. Pass `--reply-prefix [bot]`.
2. Inject processable app_mention.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = append(req.Args, "--reply-prefix", "[bot]")
	req.InjectEvents = []InjectedEvent{{
		Kind:    "app_mention",
		Channel: slackTestChannelID,
		Text:    "<@" + slackTestBotUserID + "> prefix please",
		TS:      "1710000500.000100",
	}}
	return nil
}
```