# Scenario

**Feature**: A result error line maps to one error AgentEvent

```
# result error (is_error=true or subtype=error) -> ActionError carrying result text
result-error -> WriteClaudeLine(result error) -> RawLog (1 AgentEvent "error")
```

## Preconditions
- `claude_types.FromClaude` maps an error `result` event
  (`subtype=="error"` or `is_error==true`) to exactly one `ActionError`
  AgentEvent carrying the result text.

## Steps
1. Set `req.ClaudeLines` to a single `result` error line.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClaudeLines = []string{
		`{"type":"result","subtype":"error","is_error":true,"result":"boom","duration_ms":1,"num_turns":1,"session_id":"sess"}`,
	}
	return nil
}
```
