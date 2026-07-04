# Scenario

**Feature**: info shows last three displayable messages chronologically

```
# rollout with five agent_message events (all displayable)
writeRolloutSession(5 messages) -> sessions.Info(lastN=3)

# only the last three messages appear in RecentMessages
SessionInfo.RecentMessages (len=3)
```

## Preconditions

- Five `agent_message` events with phases `commentary` are displayable.
- `req.LastN = 3`.

## Steps

1. Set `req.SessionID` to `01900006-cccc-7ccc-cccc-cccccccccccc`.
2. Write rollout with messages `msg-one` through `msg-five` in order.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "01900006-cccc-7ccc-cccc-cccccccccccc"
	req.LastN = 3
	lines := []string{
		userMessageLine("Info recent messages test"),
		agentMessageLine("msg-one"),
		agentMessageLine("msg-two"),
		agentMessageLine("msg-three"),
		agentMessageLine("msg-four"),
		agentMessageLine("msg-five"),
	}
	writeRolloutSession(t, req.CodexHome, req.SessionID,
		"2026-06-23T12:00:00.000Z", "/tmp/info-recent-project", lines...)
	return nil
}
```