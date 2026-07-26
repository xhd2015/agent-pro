# Scenario

**Feature**: log prints RUN and ASSISTANT for exec and agent_message events

```
# rollout with agent_message and exec_command function_call pair
writeRolloutSession -> sessions.PrintLog

# compact trace shows ASSISTANT greeting and RUN echo hi
terminal log with RUN and ASSISTANT labels
```

## Preconditions

- `exec_command` function_call and matching output are displayable.
- `agent_message` with phase `commentary` is displayable.

## Steps

1. Set session id `01900009-ffff-7fff-ffff-ffffffffffff`.
2. Write agent message plus exec_command lines.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "01900009-ffff-7fff-ffff-ffffffffffff"
	lines := []string{
		agentMessageLine("Hello from codex"),
	}
	lines = append(lines, execCommandLines("echo hi", "hi\n", "call_log_1")...)
	writeRolloutSession(t, req.CodexHome, req.SessionID,
		"2026-06-23T14:00:00.000Z", "/tmp/log-exec", lines...)
	return nil
}
```