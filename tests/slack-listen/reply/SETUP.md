# Scenario

**Feature**: agent output posted back to Slack thread

```
agent-run stdout -> chat.postMessage with thread_ts (and optional prefix)
```

## Preconditions

- slacktest captures PostMessage form fields.

## Steps

1. Daemon with reply-related flags.
2. Inject one processable event.
3. Assert captured PostMessage fields.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	prependListenTokens(req)
	req.Daemon = true
	req.Args = []string{"--no-require-mention"}
	req.WantAgentCalls = 1
	return nil
}
```