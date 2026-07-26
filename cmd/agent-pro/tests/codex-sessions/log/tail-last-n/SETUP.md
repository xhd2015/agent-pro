# Scenario

**Feature**: PrintLog with tail returns only the last N displayable events

```
# five agent messages in rollout transcript
writeRolloutSession with msg-one .. msg-five

# tail=2 keeps only the last two displayable blocks
PrintLog(path, w, 2)
```

## Preconditions

- Each `agent_message` event is displayable in the log pipeline.

## Steps

1. Set session id and `req.Tail = 2`.
2. Write five chronological agent messages.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "0190000a-aaaa-7aaa-aaaa-aaaaaaaaaaaa"
	req.Tail = 2
	lines := []string{
		agentMessageLine("msg-one"),
		agentMessageLine("msg-two"),
		agentMessageLine("msg-three"),
		agentMessageLine("msg-four"),
		agentMessageLine("msg-five"),
	}
	writeRolloutSession(t, req.CodexHome, req.SessionID,
		"2026-06-23T15:00:00.000Z", "/tmp/log-tail", lines...)
	return nil
}
```