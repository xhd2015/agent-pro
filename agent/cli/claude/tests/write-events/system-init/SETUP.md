# Scenario

**Feature**: A system init NDJSON line maps to one step_start AgentEvent

```
# system init line carries session_id; FromClaude emits ActionStepStart
system-init -> WriteClaudeLine(system init) -> RawLog (1 AgentEvent "step_start")
```

## Preconditions
- `claude_types.FromClaude` maps a `system` event (subtype `init`) to exactly
  one `ActionStepStart` AgentEvent.

## Steps
1. Set `req.ClaudeLines` to a single `system` init line.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClaudeLines = []string{
		`{"type":"system","subtype":"init","cwd":"/tmp","session_id":"sess","model":"claude-sonnet","tools":[],"permissionMode":"default"}`,
	}
	return nil
}
```
