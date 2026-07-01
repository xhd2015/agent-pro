# Scenario

**Feature**: A result success line maps to one done AgentEvent

```
# result success (is_error=false) -> ActionDone carrying result text
result-success -> WriteClaudeLine(result success) -> RawLog (1 AgentEvent "done")
```

## Preconditions
- `claude_types.FromClaude` maps a non-error `result` event to exactly one
  `ActionDone` AgentEvent carrying the result text.

## Steps
1. Set `req.ClaudeLines` to a single `result` success line.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClaudeLines = []string{
		`{"type":"result","subtype":"success","is_error":false,"result":"pong","duration_ms":1,"num_turns":1,"session_id":"sess","stop_reason":"end_turn"}`,
	}
	return nil
}
```
