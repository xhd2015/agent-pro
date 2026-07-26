# Scenario

**Feature**: Slack reply after agent (stateless capture vs thread open)

```
thread: interactive open -> no agent-body PostMessage
stateless: agent-run stdout -> chat.postMessage with thread_ts (+ optional prefix)
```

## Preconditions

- slacktest captures PostMessage form fields.
- Default session mode is thread unless leaf sets `--session-mode stateless`.
- Set `WantPosts` when the leaf expects PostMessage captures.

## Steps

1. Daemon with reply-related flags.
2. Inject one processable event.
3. Assert captured PostMessage fields and/or zero posts for thread open.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClearSlackEnv = true
	prependListenTokens(req)
	req.Daemon = true
	req.Args = []string{"--no-require-mention"}
	req.WantAgentCalls = 1
	return nil
}
```
