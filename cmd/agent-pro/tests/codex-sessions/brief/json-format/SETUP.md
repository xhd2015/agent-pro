# Scenario

**Feature**: brief JSON includes recent_messages array

```
# rollout with displayable agent_message events
writeRolloutSession -> sessions.Brief -> FormatBriefJSON

# JSON includes session metadata and recent_messages
{"recent_messages":[{"kind","text","formatted"}]}
```

## Preconditions

- `req.Format = "json"`.

## Steps

1. Set session id `01900007-dddd-7ddd-dddd-dddddddddddd`.
2. Write two agent messages and request brief JSON.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "01900007-dddd-7ddd-dddd-dddddddddddd"
	req.Format = "json"
	req.LastN = 3
	lines := []string{
		agentMessageLine("first insight"),
		agentMessageLine("second insight"),
	}
	writeRolloutSession(t, req.CodexHome, req.SessionID,
		"2026-06-23T13:00:00.000Z", "/tmp/json-brief", lines...)
	return nil
}
```