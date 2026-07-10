# Scenario

**Feature**: --no-require-mention processes plain channel messages

```
channel message + --no-require-mention -> agent invoked
```

## Steps

1. Pass `--no-require-mention`.
2. Inject plain channel message.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--no-require-mention"}
	req.WantAgentCalls = 1
	req.InjectEvents = []InjectedEvent{{
		Channel: slackTestChannelID,
		Text:    "channel msg no mention flag off",
		TS:      "1710000004.000100",
	}}
	return nil
}
```