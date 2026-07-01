# Scenario

**Feature**: A user tool_result line emits no AgentEvent

```
# user tool_result folds into the preceding tool call; FromClaude skips it
user-tool-result -> WriteClaudeLine(user tool_result) -> RawLog (0 AgentEvents)
```

## Preconditions
- `claude_types.FromClaude` skips `user` events entirely (tool_result echoes
  fold into the preceding tool call).

## Steps
1. Set `req.ClaudeLines` to one `user` tool_result line.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClaudeLines = []string{
		`{"type":"user","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"t1","content":"output","is_error":false}]},"session_id":"sess"}`,
	}
	return nil
}
```
