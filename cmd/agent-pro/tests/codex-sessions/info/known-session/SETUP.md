# Scenario

**Feature**: info returns status, line count, rollout path, and token totals

```
# rollout with task lifecycle, messages, and token_count events
writeRolloutSession -> sessions.Info -> FormatInfoText(now)

# output includes Session, Status, Lines, File, Recent messages, Tokens
terminal key-value text
```

## Preconditions

- Rollout includes `task_started`, displayable messages, and `token_count` events.
- `req.Now` is fixed for relative last-active time.

## Steps

1. Write a session with user message title, agent reply, and token counts.
2. Set `req.SessionID` to the fixture UUID.

```go
import "testing"

const knownInfoSessionID = "019f283a-aaaa-7aaa-aaaa-aaaaaaaaaaaa"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = knownInfoSessionID
	lines := []string{
		`{"type":"event_msg","payload":{"type":"task_started"}}`,
		userMessageLine("Refactor auth module"),
		agentMessageLine("Analyzing auth flow"),
		tokenCountLine(1200, 450),
		tokenCountLine(300, 150),
		`{"type":"event_msg","payload":{"type":"task_complete"}}`,
	}
	writeRolloutSession(t, req.CodexHome, knownInfoSessionID,
		"2026-07-03T13:00:00.000Z", "/tmp/codex-known-project", lines...)
	return nil
}
```