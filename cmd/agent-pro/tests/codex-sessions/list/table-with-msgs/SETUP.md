# Scenario

**Feature**: list table MSGS column reflects displayable event count

```
# one rollout with five displayable agent_message events
writeRolloutSession(5 messages) -> sessions.List -> FormatListTable(now)

# table header has MSGS; row shows 5
terminal table text
```

## Preconditions

- `req.Now` is fixed by root Setup.
- Five `agent_message` events with phase `commentary` are displayable.

## Steps

1. Create one session with five agent messages.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Limit = 10
	lines := []string{
		userMessageLine("Count displayable events"),
		agentMessageLine("msg-one"),
		agentMessageLine("msg-two"),
		agentMessageLine("msg-three"),
		agentMessageLine("msg-four"),
		agentMessageLine("msg-five"),
	}
	writeRolloutSession(t, req.CodexHome,
		"01900005-aaaa-7aaa-aaaa-aaaaaaaaaaaa",
		"2026-07-03T14:30:00.000Z", "/tmp/msgs-project", lines...)
	return nil
}
```